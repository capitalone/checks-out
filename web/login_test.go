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
	"testing"
	"time"

	"github.com/capitalone/checks-out/cache"
	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/store"
)

type mockStore struct {
	store.Store
	callCount map[string]int
}

func (ms *mockStore) GetValidOrgs() ([]string, error) {
	ms.callCount["getValidOrgs"]++
	return []string{"beatles", "stones", "ledzep"}, nil
}

func (ms *mockStore) CheckValidUser(login string) (bool, error) {
	ms.callCount["getValidOrgs"]++
	return login == "frankie", nil
}

type mockRemote struct {
	remote.Remote
	callCount map[string]int
}

func (mr *mockRemote) GetOrgs(c context.Context, user *model.User) ([]*model.GitHubOrg, error) {
	mr.callCount["getOrgs"]++
	if user.Login == "john" {
		return []*model.GitHubOrg{
			{
				Login: "beatles",
			},
		}, nil
	}

	if user.Login == "eric" {
		return []*model.GitHubOrg{
			{
				Login: "cream",
			},
		}, nil
	}

	return []*model.GitHubOrg{}, nil
}

func TestValidateUserAccess(t *testing.T) {
	s := &mockStore{callCount: map[string]int{}}
	r := &mockRemote{callCount: map[string]int{}}

	c := context.Background()
	c = context.WithValue(c, "store", s)
	c = context.WithValue(c, "cache", cache.NewTTL(40*time.Millisecond))
	c = context.WithValue(c, "remote", r)

	// case 1 -- no validation
	envvars.Env.Access.LimitOrgs = false
	envvars.Env.Access.LimitUsers = false

	u := &model.User{
		Login: "john",
	}
	err := validateUserAccess(c, u)
	if err != nil {
		t.Errorf("Expected no err, got %v", err)
	}

	// case 2 -- validate user only
	envvars.Env.Access.LimitUsers = true
	err = validateUserAccess(c, u)
	if err == nil {
		t.Errorf("Expected err, got %v", err)
	}

	//2a -- valid user
	u.Login = "frankie"
	err = validateUserAccess(c, u)
	if err != nil {
		t.Errorf("Expected no err, got %v", err)
	}

	// case 3 -- validate org only (frankie fails, john passes)
	envvars.Env.Access.LimitOrgs = true
	envvars.Env.Access.LimitUsers = false

	err = validateUserAccess(c, u)
	if err == nil {
		t.Errorf("Expected err, got %v", err)
	}

	//case 3a -- valid user
	u.Login = "john"
	err = validateUserAccess(c, u)
	if err != nil {
		t.Errorf("Expected no err, got %v", err)
	}

	// case 4 - valid org and user (frankie works, john works, eric fails, waldo fails)
	envvars.Env.Access.LimitOrgs = true
	envvars.Env.Access.LimitUsers = true

	u.Login = "frankie"
	err = validateUserAccess(c, u)
	if err != nil {
		t.Errorf("Expected no err, got %v", err)
	}

	u.Login = "john"
	err = validateUserAccess(c, u)
	if err != nil {
		t.Errorf("Expected no err, got %v", err)
	}

	u.Login = "eric"
	err = validateUserAccess(c, u)
	if err == nil {
		t.Errorf("Expected err, got %v", err)
	}

	u.Login = "waldo"
	err = validateUserAccess(c, u)
	if err == nil {
		t.Errorf("Expected err, got %v", err)
	}
}
