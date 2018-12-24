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

	log "github.com/Sirupsen/logrus"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
)

func isBehind(c context.Context, user *model.User, repo *model.Repo, branch model.Branch) (bool, error) {
	resp, err := remote.CompareBranches(c, user, repo, branch.BaseName, branch.CompareName, branch.CompareOwner)
	if err != nil {
		return false, err
	}
	return resp.BehindBy > 0, nil
}

func doMerge(c context.Context, user *model.User,
	hook *StatusHook, req *model.ApprovalRequest, policy *model.ApprovalPolicy, mergeMethod string) (string, error) {
	approvals, err := buildApprovers(c, user, req)
	if err != nil {
		return "", err
	}
	var people []*model.Person
	for id := range approvals.Approvers {
		if p, ok := req.Maintainer.People[id]; ok {
			people = append(people, p)
		} else {
			people = append(people, &model.Person{Login: id})
		}
	}
	message := getCommitComment(req, policy)
	log.Debugf("parsed out commit comment message, got: %v", message)

	SHA, err := remote.MergePR(c, user, hook.Repo, *req.PullRequest, people, message, mergeMethod)
	if err != nil {
		return "", err
	}
	return SHA, nil
}

func doMergeDelete(c context.Context, user *model.User,
	hook *StatusHook, req *model.ApprovalRequest) error {
	// Head branch contains what changes you like to be applied.
	// Do not delete the base branch.
	ref := req.PullRequest.Branch.CompareName
	return remote.DeleteBranch(c, user, hook.Repo, ref)
}
