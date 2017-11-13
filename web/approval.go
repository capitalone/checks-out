/*

SPDX-Copyright: Copyright (c) Capital One Services, LLC
SPDX-License-Identifier: Apache-2.0
Copyright 2017 Capital One Services, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and limitations under the License.

*/
package web

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/logstats"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/set"
)

// https://help.github.com/articles/closing-issues-via-commit-messages/
var closePullRequest = regexp.MustCompile(`(closes|closed|close|fixes|fixed|fix|resolves|resolved|resolve) #(\d+)`)

type CurCommentStatus int

const (
	CurCommentOther CurCommentStatus = iota
	CurCommentNoChange
	CurCommentApproval
	CurCommentDisapproval
	CurCommentPRAuthor
	CurCommentPRTitle
	CurCommentPRAudit
)

type CurCommentInfo struct {
	Status CurCommentStatus
	Author string
}

type ApprovalInfo struct {
	Policy         *model.ApprovalPolicy
	Approved       bool
	AuthorApproved bool
	TitleApproved  bool
	AuditApproved  bool
	Approvers      set.Set
	Disapprovers   set.Set
	CurCommentInfo
}

type PullRequestFeedback struct {
	All         []model.Feedback
	Approval    []model.Feedback
	Disapproval []model.Feedback
}

func approve(c context.Context, params HookParams, id int, setStatus bool) (*ApprovalInfo, error) {
	user := params.User
	repo := params.Repo

	pullRequest, err := remote.GetPullRequest(c, user, repo, id)
	if err != nil {
		return nil, err
	}

	return approvePullRequest(c, params, id, &pullRequest, setStatus)
}

func approvePullRequest(c context.Context, params HookParams, id int, pullRequest *model.PullRequest, setStatus bool) (*ApprovalInfo, error) {
	user := params.User
	repo := params.Repo
	config := params.Config
	maintainer := params.Snapshot

	request := model.ApprovalRequest{
		Repository:  repo,
		Config:      config,
		Maintainer:  maintainer,
		PullRequest: pullRequest,
	}

	approval, err := buildApprovers(c, user, &request)
	if err != nil {
		return nil, err
	}

	if setStatus {
		status, desc := generateStatus(approval)

		err = remote.SetStatus(c, user, repo, pullRequest.Branch.CompareSHA, model.ServiceName, status, desc)
		if err != nil {
			return nil, err
		}

		recordStats(approval, repo, id)
	}

	log.Debugf("processed comment for %s. received %d approvals and %d disapprovals",
		repo.Slug, len(approval.Approvers), len(approval.Disapprovers))

	return approval, nil

}

func getFeedbackRanges(c context.Context,
	user *model.User,
	request *model.ApprovalRequest,
	policy *model.ApprovalPolicy) (PullRequestFeedback, error) {
	var err error
	config := request.Config
	fb := PullRequestFeedback{}
	fb.All, err = getFeedback(c, user, request, policy, model.All, config.Commit.IgnoreUIMerge)
	if err != nil {
		return fb, err
	}
	if config.Commit.Range == model.All {
		fb.Approval = fb.All
	} else {
		fb.Approval, err = getFeedback(c, user, request, policy, config.Commit.Range, config.Commit.IgnoreUIMerge)
	}
	if err != nil {
		return fb, err
	}
	if config.Commit.Range == config.Commit.AntiRange {
		fb.Disapproval = fb.Approval
	} else {
		fb.Disapproval, err = getFeedback(c, user, request, policy, config.Commit.AntiRange, config.Commit.IgnoreUIMerge)
	}
	return fb, err
}

func getIssuesFromMessage(c context.Context, user *model.User, repo *model.Repo,
	message string, numbers set.Set, issues []*model.Issue) []*model.Issue {
	message = strings.ToLower(message)
	comments := closePullRequest.FindAllStringSubmatch(message, -1)
	for _, comment := range comments {
		if len(comment) != 3 {
			log.Errorf("Issue close does not have 3 parts: %v", comment)
			continue
		}
		numbers.Add(comment[2])
	}
	for _, number := range numbers.Keys() {
		num, err := strconv.Atoi(number)
		if err != nil {
			log.Errorf("Unable to convert match (\\d+) into number: %s", number)
			continue
		}
		i, err := remote.GetIssue(c, user, repo, num)
		if err != nil {
			log.Warnf("Unable to fetch issue %s/%s/%d", repo.Owner, repo.Name, num)
			continue
		}
		issues = append(issues, &i)
	}
	return issues
}

func getIssues(c context.Context, u *model.User, r *model.Repo,
	pr *model.PullRequest, feedback []model.Feedback) []*model.Issue {
	numbers := set.Empty()
	issues := getIssuesFromMessage(c, u, r, pr.Body, numbers, nil)
	for _, fb := range feedback {
		issues = getIssuesFromMessage(c, u, r, fb.GetBody(), numbers, issues)
	}
	return issues
}

func buildApprovers(c context.Context,
	user *model.User,
	request *model.ApprovalRequest) (*ApprovalInfo, error) {
	repo := request.Repository
	pr := request.PullRequest
	files, err := getPullRequestFiles(c, user, repo, pr.Number)
	if err != nil {
		return nil, err
	}
	request.Files = files
	policy := model.FindApprovalPolicy(request)
	feedback, err := getFeedbackRanges(c, user, request, policy)
	if err != nil {
		return nil, err
	}
	request.ApprovalComments = feedback.Approval
	request.DisapprovalComments = feedback.Disapproval
	request.Issues = getIssues(c, user, repo, pr, feedback.All)
	audit, err := calculateAuditInfo(c, user, request)
	if err != nil {
		return nil, err
	}
	return calculateApprovalInfo(request, policy, audit), nil
}

func calculateAuditInfo(c context.Context, user *model.User, request *model.ApprovalRequest) (bool, error) {
	var validAudit bool
	var err error
	if requireAudit(request.Config, request.PullRequest) {
		validAudit, err = testAudit(c, user, request.Repository, request.PullRequest)
	} else {
		validAudit = true
	}
	return validAudit, err
}

func calculateApprovalInfo(request *model.ApprovalRequest, policy *model.ApprovalPolicy, audit bool) *ApprovalInfo {
	approvers := set.Empty()
	disapprovers := set.Empty()
	validAuthor := false
	validTitle := false
	validAudit := audit
	approved := model.Approve(request, policy,
		func(f model.Feedback, op model.ApprovalOp) {
			author := f.GetAuthor().String()
			switch op {
			case model.Approval:
				approvers.Add(author)
			case model.DisapprovalInsert:
				disapprovers.Add(author)
			case model.DisapprovalRemove:
				disapprovers.Remove(author)
			case model.ValidAuthor:
				validAuthor = true
			case model.ValidTitle:
				validTitle = true
			default:
				panic(fmt.Sprintf("Unknown approval operation %d", op))
			}
		})
	approved = approved && audit

	ai := ApprovalInfo{
		Policy:         policy,
		Approved:       approved,
		AuthorApproved: validAuthor,
		TitleApproved:  validTitle,
		AuditApproved:  validAudit,
		Approvers:      approvers,
		Disapprovers:   disapprovers,
		CurCommentInfo: CurCommentInfo{
			Author: "",
			Status: CurCommentNoChange,
		},
	}

	if !validAudit {
		ai.Status = CurCommentPRAudit
		// need to check title before author, since it's processed first and
		// if title is triggered, then author never gets processed
	} else if !validTitle {
		ai.Status = CurCommentPRTitle
	} else if !validAuthor {
		ai.Status = CurCommentPRAuthor
	} else if len(request.ApprovalComments) > 0 && len(request.DisapprovalComments) > 0 {
		//go back and check if the last comment is an approval or a block
		lookback := *request
		lookback.ApprovalComments = request.ApprovalComments[len(request.ApprovalComments)-1:]
		lookback.DisapprovalComments = request.DisapprovalComments[len(request.DisapprovalComments)-1:]
		model.Approve(&lookback, policy,
			func(f model.Feedback, op model.ApprovalOp) {
				switch op {
				case model.Approval:
					ai.Status = CurCommentApproval
					ai.Author = f.GetAuthor().String()
				case model.DisapprovalInsert:
					ai.Status = CurCommentDisapproval
					ai.Author = f.GetAuthor().String()
				}
			})
	}

	return &ai
}

func generateStatus(info *ApprovalInfo) (string, string) {
	status := "pending"
	var desc string
	if info.Approved {
		status = "success"
		if len(info.Approvers) > 0 {
			desc = "approved by " + info.Approvers.Print(",")
		} else {
			desc = "approval did not require approvers"
		}
	} else if !info.AuditApproved {
		status = "error"
		desc = "audit chain must be manually approved"
	} else if !info.TitleApproved {
		// put title first because we check title first
		// longer term fix is to get rid of the side effects
		status = "error"
		desc = "pull request title is blocking merge"
	} else if !info.AuthorApproved {
		status = "error"
		desc = "pull request author not allowed"
	} else if len(info.Disapprovers) > 0 {
		desc = "blocked by " + info.Disapprovers.Print(",")
	} else if len(info.Approvers) > 0 {
		desc = fmt.Sprintf("more approvals needed. %s: %s", envvars.Env.Branding.ShortName, info.Approvers.Print(","))
	} else {
		desc = "no approvals received"
	}
	return status, desc
}

func recordStats(approval *ApprovalInfo, repo *model.Repo, pr int) {
	if approval.Approved && len(approval.Approvers) > 0 {
		id := fmt.Sprintf("%s/%s/%d", repo.Owner, repo.Name, pr)
		logstats.RecordPR(id)
	}
	for id := range approval.Approvers {
		logstats.RecordApprover(id)
	}
	for id := range approval.Disapprovers {
		logstats.RecordDisapprover(id)
	}
}
