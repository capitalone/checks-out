/*

SPDX-Copyright: Copyright (c) Brad Rydzewski, project contributors, Capital One Services, LLC
SPDX-License-Identifier: Apache-2.0
Copyright 2017 Brad Rydzewski, project contributors, Capital One Services, LLC

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

	"github.com/gin-gonic/gin"
	"github.com/capitalone/checks-out/model"
)

type Hook interface {
	Process(c context.Context) (interface{}, error)
}

type ApprovalHook struct {
	Repo  *model.Repo
	Issue *model.Issue
}

type CommentHook struct {
	ApprovalHook
	Comment string
}

type ReviewHook struct {
	ApprovalHook
}

type PRHook struct {
	ApprovalHook
	PullRequest *model.PullRequest
	Action      string
}

type RepoHook struct {
	Name    string
	Owner   string
	Action  string
	BaseURL string
}

type StatusHook struct {
	SHA    string
	Status *model.CommitStatus
	Repo   *model.Repo
}

type HookParams struct {
	Repo     *model.Repo
	User     *model.User
	Cap      *model.Capabilities
	Config   *model.Config
	Snapshot *model.MaintainerSnapshot
}

func ProcessHook(c *gin.Context) {
	hook, c2, err := createHook(c, c.Request)
	if err != nil {
		c.Error(err)
	} else if hook == nil {
		c.String(200, "pong")
	} else {
		output, err := hook.Process(c2)
		if err != nil {
			c.Error(err)
		} else {
			if output == nil {
				c.String(200, "pong")
			} else {
				c.IndentedJSON(200, output)
			}

		}
	}
}
