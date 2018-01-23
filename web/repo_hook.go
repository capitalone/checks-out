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

	"github.com/capitalone/checks-out/api"
	"github.com/capitalone/checks-out/store"
)

func (hook *RepoHook) Process(c context.Context) (interface{}, error) {
	return doRepoHook(c, hook)
}

type RepoOutput struct {
	Action string `json:"action"`
	Owner  string `json:"owner"`
	Name   string `json:"name"`
}

func doRepoHook(c context.Context, hook *RepoHook) (*RepoOutput, error) {

	orgDb, err := store.GetOrgName(c, hook.Owner)
	if err != nil {
		return nil, err
	}
	user, err := store.GetUser(c, orgDb.UserID)
	if err != nil {
		return nil, err
	}

	repoOutput := &RepoOutput{
		Action: hook.Action,
		Owner:  hook.Owner,
		Name:   hook.Name,
	}

	switch hook.Action {
	case "created":
		api.TurnOnRepoQuiet(c, user, hook.Owner, hook.Name, hook.BaseURL)
	case "deleted":
		repo, err := store.GetRepoOwnerName(c, hook.Owner, hook.Name)
		if err != nil {
			api.TurnOffRepo(c, user, repo, hook.Owner, hook.Name, hook.BaseURL)
		}
	default:
		repoOutput = nil
	}
	return repoOutput, nil
}
