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
	"sort"
	"strings"

	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
)

func getPullRequestFiles(c context.Context, user *model.User, repo *model.Repo, num int) ([]model.CommitFile, error) {
	files, err := remote.GetPullRequestFiles(c, user, repo, num)
	if err != nil {
		msg := fmt.Sprintf("Error retrieving files for %s pr %d", repo.Slug, num)
		err = exterror.Append(err, msg)
		return nil, err
	}
	return files, nil
}

func getPullRequestCommits(c context.Context, user *model.User, repo *model.Repo, num int) ([]model.Commit, error) {
	commits, err := remote.GetPullRequestCommits(c, user, repo, num)
	if err != nil {
		msg := fmt.Sprintf("Error retrieving commits for %s pr %d", repo.Slug, num)
		err = exterror.Append(err, msg)
		return nil, err
	}
	return commits, nil
}

func getFeedback(c context.Context,
	user *model.User,
	request *model.ApprovalRequest,
	policy *model.ApprovalPolicy,
	crange model.CommitRange,
	noUIMerge bool) ([]model.Feedback, error) {
	repo := request.Repository
	pr := request.PullRequest
	var feedback []model.Feedback
	var err error
	fbConfig := request.Config.GetFeedbackConfig(policy)
	switch crange {
	case model.All:
		feedback, err = getAllFeedback(c, user, repo, pr.Number, fbConfig.Types)
	case model.Head:
		feedback, err = getFeedbackSinceHead(c, user, repo, pr.Number, noUIMerge, fbConfig.Types)
	default:
		feedback, err = nil, fmt.Errorf("Unknown commit range '%s' in configuration",
			crange.String())
	}
	if err != nil {
		return nil, err
	}
	return feedback, nil
}

func hasFeedbackType(types []model.FeedbackType, target model.FeedbackType) bool {
	for _, t := range types {
		if t == target {
			return true
		}
	}
	return false
}

func filterFeedback(feedback []model.Feedback) []model.Feedback {
	filter := feedback[:0]
	for _, msg := range feedback {
		if !strings.HasPrefix(msg.GetBody(), model.CommentPrefix) {
			filter = append(filter, msg)
		}
	}
	return filter
}

func sortFeedback(feedback []model.Feedback) {
	sort.SliceStable(feedback, func(i, j int) bool {
		return feedback[i].GetSubmittedAt().Before(feedback[j].GetSubmittedAt())
	})
}

func getAllFeedback(c context.Context, user *model.User, repo *model.Repo, num int, types []model.FeedbackType) ([]model.Feedback, error) {
	var feedback []model.Feedback
	var comments []*model.Comment
	var reviews []*model.Review
	var err error
	if hasFeedbackType(types, model.CommentType) {
		comments, err = remote.GetAllComments(c, user, repo, num)
		if err != nil {
			msg := fmt.Sprintf("Error retrieving comments for %s pr %d", repo.Slug, num)
			err = exterror.Append(err, msg)
			return nil, err
		}
	}
	if hasFeedbackType(types, model.ReviewType) {
		reviews, err = remote.GetAllReviews(c, user, repo, num)
		if err != nil {
			msg := fmt.Sprintf("Error retrieving reviews for %s pr %d", repo.Slug, num)
			err = exterror.Append(err, msg)
			return nil, err
		}
	}
	for _, c := range comments {
		feedback = append(feedback, c)
	}
	for _, r := range reviews {
		feedback = append(feedback, r)
	}
	feedback = filterFeedback(feedback)
	sortFeedback(feedback)
	return feedback, nil
}

func getFeedbackSinceHead(c context.Context, user *model.User, repo *model.Repo, num int, noUIMerge bool, types []model.FeedbackType) ([]model.Feedback, error) {
	var feedback []model.Feedback
	var comments []*model.Comment
	var reviews []*model.Review
	var err error
	if hasFeedbackType(types, model.CommentType) {
		comments, err = remote.GetCommentsSinceHead(c, user, repo, num, noUIMerge)
		if err != nil {
			msg := fmt.Sprintf("Error retrieving comments for %s pr %d", repo.Slug, num)
			err = exterror.Append(err, msg)
			return nil, err
		}
	}
	if hasFeedbackType(types, model.ReviewType) {
		reviews, err = remote.GetReviewsSinceHead(c, user, repo, num, noUIMerge)
		if err != nil {
			msg := fmt.Sprintf("Error retrieving reviews for %s pr %d", repo.Slug, num)
			err = exterror.Append(err, msg)
			return nil, err
		}
	}
	for _, c := range comments {
		feedback = append(feedback, c)
	}
	for _, r := range reviews {
		feedback = append(feedback, r)
	}
	feedback = filterFeedback(feedback)
	sortFeedback(feedback)
	return feedback, nil
}

func sendErrorStatusPR(c context.Context, hook *ApprovalHook, pr *model.PullRequest, e error) error {
	repo, user, _, err := GetRepoAndUser(c, hook.Repo.Slug)
	if err != nil {
		return err
	}
	return remote.SetStatus(c, user, repo, pr.Branch.CompareSHA, model.ServiceName, "error", e.Error())
}

func sendErrorStatus(c context.Context, hook *ApprovalHook, e error) error {
	repo, user, _, err := GetRepoAndUser(c, hook.Repo.Slug)
	if err != nil {
		return err
	}
	pr, err := remote.GetPullRequest(c, user, repo, hook.Issue.Number)
	if err != nil {
		return err
	}
	return remote.SetStatus(c, user, repo, pr.Branch.CompareSHA, model.ServiceName, "error", e.Error())
}
