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

	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/snapshot"
	"github.com/capitalone/checks-out/store"
)

func GetHookParameters(c context.Context, h HookCommon, slug string) (HookParams, error) {
	result, err := GetHookParametersBasic(c, slug)
	if err != nil {
		return HookParams{}, err
	}
	err = snapshot.FixSlackTargets(c, result.Config, result.User.Login)
	if err != nil {
		return HookParams{}, err
	}
	result.Event = h.Event()
	result.Action = h.Action()
	return result, nil
}

func GetHookParametersBasic(c context.Context, slug string) (HookParams, error) {
	repo, user, cap, err := GetRepoAndUser(c, slug)
	if err != nil {
		return HookParams{}, err
	}
	config, maintainer, err := snapshot.GetConfigAndMaintainers(c, user, cap, repo)
	if err != nil {
		return HookParams{}, err
	}
	result := HookParams{
		Repo:     repo,
		User:     user,
		Cap:      cap,
		Config:   config,
		Snapshot: maintainer,
	}
	return result, nil
}

func GetRepoAndUser(c context.Context, slug string) (*model.Repo, *model.User, *model.Capabilities, error) {
	repo, err := store.GetRepoSlug(c, slug)
	if err != nil {
		msg := fmt.Sprintf("Error getting repository %s", slug)
		err = exterror.Append(err, msg)
		return nil, nil, nil, err
	}
	user, err := store.GetUser(c, repo.UserID)
	if err != nil {
		msg := fmt.Sprintf("Error getting repository owner %s", repo.Slug)
		err = exterror.Append(err, msg)
		return nil, nil, nil, err
	}
	cap, err := remote.Capabilities(c, user)
	if err != nil {
		msg := fmt.Sprintf("Error getting repository %s", slug)
		err = exterror.Append(err, msg)
		return nil, nil, nil, err
	}
	return repo, user, cap, err
}
