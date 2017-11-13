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

	log "github.com/Sirupsen/logrus"
	multierror "github.com/mspiegel/go-multierror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/notifier"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/set"
)

type ApprovalOutput struct {
	Policy       *model.ApprovalPolicy `json:"policy"`
	Settings     *model.Config         `json:"config"`
	Approved     bool                  `json:"approved"`
	Approvers    set.Set               `json:"approvers"`
	Disapprovers set.Set               `json:"disapprovers"`
}

var actionWhiteList = set.New("synchronize", "opened", "reopened", "closed")

func (hook *PRHook) Process(c context.Context) (interface{}, error) {
	approvalOutput, e1 := doPRHookAndNotify(c, hook)
	if e1 != nil {
		e2 := sendErrorStatusPR(c, &hook.ApprovalHook, hook.PullRequest, e1)
		e1 = multierror.Append(e1, e2)
	}
	return approvalOutput, e1
}

func doPRHookAndNotify(c context.Context, hook *PRHook) (*ApprovalOutput, error) {
	if !actionWhiteList.Contains(hook.Action) {
		return nil, nil
	}
	params, err := GetHookParameters(c, hook.Repo.Slug, true)
	if err != nil {
		return nil, err
	}
	approvalOutput, mw, err := doPRHook(c, hook, params)
	mw = createMessage(hook, mw, err)
	if mw != nil {
		notifier.SendMessage(c, params.Config, *mw)
	}
	return approvalOutput, err
}

func doPRHook(c context.Context, hook *PRHook, params HookParams) (*ApprovalOutput, *notifier.MessageWrapper, error) {
	switch hook.Action {
	case "closed":
		mw, err := prClosed(c, hook, params)
		return nil, mw, err
	case "opened", "reopened":
		err := prOpened(c, hook, params)
		if err != nil {
			return nil, nil, err
		}
	}
	return prHandle(c, hook, params)
}

func prOpened(c context.Context, hook *PRHook, params HookParams) error {
	request := model.ApprovalRequest{
		Repository:  params.Repo,
		Config:      params.Config,
		PullRequest: hook.PullRequest,
	}
	success, err := calculateAuditInfo(c, params.User, &request)
	if err != nil {
		return err
	}
	if success {
		return nil
	}
	return manualAudit(c, params.User, params.Repo, hook.PullRequest)
}

func prHandle(c context.Context, hook *PRHook, params HookParams) (*ApprovalOutput, *notifier.MessageWrapper, error) {
	var approvalOutput *ApprovalOutput
	approvalInfo, err := approvePullRequest(c, params, hook.Issue.Number, hook.PullRequest, true)

	if err == nil {
		approvalOutput = &ApprovalOutput{
			Policy:       approvalInfo.Policy,
			Settings:     params.Config,
			Approved:     approvalInfo.Approved,
			Approvers:    approvalInfo.Approvers,
			Disapprovers: approvalInfo.Disapprovers,
		}
	}

	mw := handleNotification(c, hook, params, approvalInfo)

	if err == nil {
		mw.Merge(*handleApprovalNotification(&hook.ApprovalHook, &approvalInfo.CurCommentInfo))
	}
	return approvalOutput, mw, err
}

func prClosed(c context.Context, hook *PRHook, params HookParams) (*notifier.MessageWrapper, error) {
	user := params.User
	repo := params.Repo
	config := params.Config

	pr, err := remote.GetPullRequest(c, user, repo, hook.Issue.Number)
	if err != nil {
		return nil, err
	}

	if requireAudit(config, &pr) {
		audit, err := testAudit(c, user, repo, &pr)
		if err != nil {
			return nil, err
		}
		if audit {
			err = applyAudit(c, user, repo, &pr)
			if err != nil {
				return nil, err
			}
		}
	}
	mw := handleNotification(c, hook, params, nil)
	return mw, nil
}

func policyDescription(ai *ApprovalInfo) string {
	if ai == nil {
		return ""
	}
	if len(ai.Policy.Name) > 0 {
		return ai.Policy.Name
	}
	if ai.Policy.Position > 0 {
		return fmt.Sprintf("# %d", ai.Policy.Position)
	}
	return ""
}

func handleNotification(c context.Context, prHook *PRHook, params HookParams, ai *ApprovalInfo) *notifier.MessageWrapper {
	mw := &notifier.MessageWrapper{
		MessageHeader: notifier.MessageHeader{
			PrName:   prHook.Issue.Title,
			PrNumber: prHook.Issue.Number,
			Slug:     prHook.Repo.Slug,
		},
	}
	if ai == nil {
		return mw
	}
	var mi notifier.MessageInfo
	switch prHook.Action {
	case "opened":
		desc := policyDescription(ai)
		mi.Message = "opened"
		if len(desc) > 0 {
			mi.Message += fmt.Sprintf(" Applying approval policy %s", desc)
		}
		mi.Type = model.CommentOpen
	case "closed":
		if prHook.PullRequest.Branch.Merged {
			mi.Message = "merged"
			mi.Type = model.CommentAccept
		} else {
			mi.Message = "closed without being merged"
			mi.Type = model.CommentClose
		}
	case "reopened":
		mi.Message = "reopened"
		mi.Type = model.CommentOpen
	case "synchronize":
		if params.Config.Commit.Range == model.Head {
			if params.Config.Commit.IgnoreUIMerge {
				merge, err := remote.IsHeadUIMerge(c, params.User, prHook.Repo, prHook.Issue.Number)
				if err != nil {
					log.Warnf("Unable to test HEAD of pull request %s/%s/%d",
						prHook.Repo.Owner, prHook.Repo.Name, prHook.Issue.Number)
				} else if merge {
					mi.Message = "merged through the user interface. Merge commit ignored."
					mi.Type = model.CommentPushIgnore
				} else {
					mi.Message = "updated. No comments before this one will count for approval."
					mi.Type = model.CommentReset
				}
			} else {
				mi.Message = "updated. No comments before this one will count for approval."
				mi.Type = model.CommentReset
			}
		}
	}
	if mi.Message != "" {
		mw.Messages = append(mw.Messages, mi)
	}
	return mw
}

func createMessage(hook *PRHook, mw *notifier.MessageWrapper, err error) *notifier.MessageWrapper {
	if mw == nil && err == nil {
		return nil
	}
	if err == nil {
		return mw
	}
	mw2 := notifier.BuildErrorMessage(hook.Issue.Title, hook.Issue.Number, hook.Repo.Slug, err.Error())
	if mw == nil {
		return &mw2
	}
	mw.Merge(mw2)
	return mw
}
