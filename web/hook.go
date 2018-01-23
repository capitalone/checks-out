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

	"github.com/capitalone/checks-out/model"
	"github.com/gin-gonic/gin"
)

type Hook interface {
	Process(c context.Context) (interface{}, error)
	SetEvent(event string)
}

type HookCommon struct {
	Event  string
	Action string
}

type ApprovalHook struct {
	HookCommon
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
}

type RepoHook struct {
	HookCommon
	Name    string
	Owner   string
	BaseURL string
}

type StatusHook struct {
	HookCommon
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
	Event    string
	Action   string
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

func (h *HookCommon) SetEvent(event string) {
	h.Event = event
}
