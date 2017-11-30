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
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/notifier"
	"github.com/capitalone/checks-out/remote"
)

type StatusResponse struct {
	SHA  string `json:"sha,omitempty"`
	Tag  string `json:"tag,omitempty"`
	Err  string `json:"error,omitempty"`
	Info string `json:"info,omitempty"`
}

func generateError(msg string, err error, v model.PullRequest, slug string,
	resp *StatusResponse, mw *notifier.MessageWrapper) {
	log.Warnf("%s: %s", msg, err)
	resp.Err = err.Error()
	mw.Merge(notifier.BuildErrorMessage(v.Issue.Title, v.Issue.Number, slug, err.Error()))
}

func (hook *StatusHook) Process(c context.Context) (interface{}, error) {

	if hook.Status.State != "success" {
		return nil, nil
	}

	params, err := GetHookParameters(c, hook.Repo.Slug, true)
	if err != nil {
		return nil, err
	}
	repo := params.Repo
	user := params.User
	config := params.Config
	maintainer := params.Snapshot

	merged := map[string]StatusResponse{}

	log.Debug("calling getPullRequestsForCommit for sha", hook.SHA)
	pullRequests, err := remote.GetPullRequestsForCommit(c, user, hook.Repo, &hook.SHA)
	log.Debugf("sha for commit is %s, pull requests are: %v", hook.SHA, pullRequests)

	if err != nil {
		notifier.SendErrorMessage(c, config, "Unknown", 0, hook.Repo.Slug, err.Error())
		return nil, err
	}

	//check the statuses of all of the checks on the branches for this commit
	for _, v := range pullRequests {
		mw := &notifier.MessageWrapper{
			MessageHeader: notifier.MessageHeader{
				PrName:   v.Issue.Title,
				PrNumber: v.Issue.Number,
				Slug:     hook.Repo.Slug,
			},
		}

		id := fmt.Sprintf("%d", v.Number)
		//if all of the statuses are success, then merge and create a tag for the version
		if v.Branch.Mergeable {
			result := StatusResponse{}

			files, err := getPullRequestFiles(c, user, hook.Repo, v.Number)
			if err != nil {
				generateError("Unable to get pull request files", err, v, hook.Repo.Slug, &result, mw)
				merged[id] = result
				sendMessage(c, config, mw)
				continue
			}

			req := &model.ApprovalRequest{
				Config:      config,
				Maintainer:  maintainer,
				PullRequest: &v,
				Repository:  repo,
				Files:       files,
			}

			policy := model.FindApprovalPolicy(req)
			mergeConfig := req.Config.GetMergeConfig(policy)

			if !mergeConfig.Enable {
				result.Info = "merge config not enabled"
				merged[id] = result
				sendMessage(c, config, mw)
				continue
			}

			success, err := remote.HasRequiredStatus(c, user, hook.Repo, v.Branch.BaseName, v.Branch.CompareSHA)

			if err != nil {
				generateError("Unable to test commit statuses", err, v, hook.Repo.Slug, &result, mw)
				merged[id] = result
				sendMessage(c, config, mw)
				continue
			}

			if !success {
				result.Info = "required status checks are not passed"
				merged[id] = result
				sendMessage(c, config, mw)
				continue
			}

			if mergeConfig.UpToDate {
				behind, err2 := isBehind(c, user, repo, v.Branch)
				if err2 != nil {
					generateError("Unable to compare branches", err2, v, hook.Repo.Slug, &result, mw)
					merged[id] = result
					sendMessage(c, config, mw)
					continue
				}
				if behind {
					result.Info = "compare branch is behind base branch"
					merged[id] = result
					sendMessage(c, config, mw)
					continue
				}
			}

			SHA, err := doMerge(c, user, hook, req, policy, mergeConfig.Method)

			if err != nil {
				generateError("Unable to merge pull request", err, v, hook.Repo.Slug, &result, mw)
				merged[id] = result
				sendMessage(c, config, mw)
				continue
			}

			mw.Messages = append(mw.Messages, notifier.MessageInfo{
				Message: "merged",
				Type:    model.CommentMerge,
			})

			result.SHA = SHA

			tag, err := tagIfEnabled(c, user, hook, req, policy, SHA)
			if err != nil {
				generateError("Unable to tag", err, v, hook.Repo.Slug, &result, mw)
				merged[id] = result
				sendMessage(c, config, mw)
				continue
			}

			result.Tag = tag

			if tag != "" {
				mw.Messages = append(mw.Messages, notifier.MessageInfo{
					Message: fmt.Sprintf("Tag %s has been added", tag),
					Type:    model.CommentTag,
				})
			}

			if mergeConfig.Delete && req.PullRequest.Branch.CompareOwner == hook.Repo.Owner {
				err = doMergeDelete(c, user, hook, req)
				if err != nil {
					generateError("Unable to delete merged branch", err, v, hook.Repo.Slug, &result, mw)
					merged[id] = result
					sendMessage(c, config, mw)
					continue
				}
				mw.Messages = append(mw.Messages, notifier.MessageInfo{
					Message: fmt.Sprintf("Branch %s has been deleted", req.PullRequest.Branch.CompareName),
					Type:    model.CommentDelete,
				})
			}

			if config.Deployment.Enable {
				doDeployment(c, user, config, hook, v.Branch.BaseName)
				mw.Messages = append(mw.Messages, notifier.MessageInfo{
					Message: fmt.Sprintf("Deployment has been triggered from branch %s", v.Branch.BaseName),
					Type:    model.CommentDeployment,
				})
			}
			merged[id] = result
			sendMessage(c, config, mw)
		}
	}
	log.Debugf("processed status for %s. received %v ", repo.Slug, hook)

	return merged, nil
}

func sendMessage(c context.Context, config *model.Config, mw *notifier.MessageWrapper) {
	notifier.SendMessage(c, config, *mw)
}

func doDeployment(c context.Context, user *model.User, config *model.Config, hook *StatusHook, baseName string) error {
	var err error
	if dc, ok := config.Deployment.DeploymentMap[baseName]; ok {
		//if we are
		env := ""
		if dc.Environment != nil {
			env = *dc.Environment
		}
		if len(dc.Tasks) == 0 && env != "" {
			scheduleDeployment(c, user, hook.Repo, baseName, nil, env)
		} else {
			for _, task := range dc.Tasks {
				scheduleDeployment(c, user, hook.Repo, baseName, &task, env)
			}
		}
	}
	return err
}

func scheduleDeployment(c context.Context, user *model.User, repo *model.Repo, baseName string, task *string, env string) {
	di := model.DeploymentInfo{
		Ref:         baseName,
		Task:        *task,
		Environment: env,
	}
	err := remote.ScheduleDeployment(c, user, repo, di)
	if err != nil {
		log.Warnf("Unable to schedule deployment %v: %s", di, err)
	}
}
