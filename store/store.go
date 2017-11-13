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
package store

import (
	"context"
	"path"

	"github.com/capitalone/checks-out/cache"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/set"
)

// Store defines a data storage abstraction for managing structured data
// in the system.
type Store interface {
	// GetMigrations gets the set of migrations that were applied when service started
	GetMigrations() set.Set

	// GetUser gets a user by unique ID.
	GetUser(int64) (*model.User, error)

	// GetUserLogin gets a user by unique Login name.
	GetUserLogin(string) (*model.User, error)

	// CreateUser creates a new user account.
	CreateUser(*model.User) error

	// UpdateUser updates a user account.
	UpdateUser(*model.User) error

	// DeleteUser deletes a user account.
	DeleteUser(*model.User) error

	// GetAllUsers gets a list of all users.
	GetAllUsers() ([]*model.User, error)

	// GetRepo gets a repo by unique ID.
	GetRepo(int64) (*model.Repo, error)

	// GetRepoSlug gets a repo by its full name.
	GetRepoSlug(string) (*model.Repo, error)

	// GetRepoMulti gets a list of multiple repos by their full name.
	GetRepoMulti(...string) ([]*model.Repo, error)

	// GetAllRepos gets a list of all repositories.
	GetAllRepos() ([]*model.Repo, error)

	// GetRepoUserId gets a list by user unique ID.
	GetRepoUserId(int64) ([]*model.Repo, error)

	// CreateRepo creates a new repository.
	CreateRepo(*model.Repo) error

	// UpdateRepo updates a user repository.
	UpdateRepo(*model.Repo) error

	// DeleteRepo deletes a user repository.
	DeleteRepo(*model.Repo) error

	// CheckValidUser returns true if the user exists in the valid_user table
	CheckValidUser(string) (bool, error)

	// GetValidOrgs returns a list of the entries in the valid_orgs table
	GetValidOrgs() ([]string, error)

	// GetUserEnabledOrgs the auto-enrolling orgs with the specified names
	GetUserEnabledOrgs(names []string) ([]*model.OrgDb, error)

	// Returns the auto-enrolling org with the specified name
	GetOrgByName(owner string) (*model.OrgDb, error)

	// CreateOrg creates a new auto-enrolling org.
	CreateOrg(org *model.OrgDb) error

	// DeleteOrg deletes an auto-enrolling org.
	DeleteOrg(org *model.OrgDb) error

	// GetReposForOrg returns all repos in the specified org
	GetReposForOrg(owner string) ([]*model.Repo, error)

	// Returns the slack URL for the specified hostname and user
	// if the user string is blank, the default (admin-level) hostname is returned
	// if no hostname is found, an empty string is returned
	GetSlackUrl(hostname string, user string) (string, error)

	// Stores or updates the slack URL for the specified hostname and user
	// if the user string is blank, the default (admin-level) hostname is stored or updated
	AddUpdateSlackUrl(hostname string, user string, url string) error

	// Deletes the slack URL for the specified hostname and user
	// if the user string is blank, the default (admin-level) hostname is deleted
	DeleteSlackUrl(hostname string, user string) error
}

// GetUser gets a user by unique ID.
func GetUser(c context.Context, id int64) (*model.User, error) {
	return FromContext(c).GetUser(id)
}

// GetUserLogin gets a user by unique Login name.
func GetUserLogin(c context.Context, login string) (*model.User, error) {
	return FromContext(c).GetUserLogin(login)
}

// CreateUser creates a new user account.
func CreateUser(c context.Context, user *model.User) error {
	return FromContext(c).CreateUser(user)
}

// UpdateUser updates a user account.
func UpdateUser(c context.Context, user *model.User) error {
	return FromContext(c).UpdateUser(user)
}

// DeleteUser deletes a user account.
func DeleteUser(c context.Context, user *model.User) error {
	return FromContext(c).DeleteUser(user)
}

// GetAllUsers gets a list of all users.
func GetAllUsers(c context.Context) ([]*model.User, error) {
	return FromContext(c).GetAllUsers()
}

// GetRepo gets a repo by unique ID.
func GetRepo(c context.Context, id int64) (*model.Repo, error) {
	return FromContext(c).GetRepo(id)
}

// GetRepoSlug gets a repo by its full name.
func GetRepoSlug(c context.Context, slug string) (*model.Repo, error) {
	return FromContext(c).GetRepoSlug(slug)
}

// GetRepoMulti gets a list of multiple repos by their full name.
func GetRepoMulti(c context.Context, slug ...string) ([]*model.Repo, error) {
	return FromContext(c).GetRepoMulti(slug...)
}

// GetAllRepos gets a list of all repositories.
func GetAllRepos(c context.Context) ([]*model.Repo, error) {
	return FromContext(c).GetAllRepos()
}

// GetRepoOwnerName gets a repo by its owner and name.
func GetRepoOwnerName(c context.Context, owner, name string) (*model.Repo, error) {
	return GetRepoSlug(c, path.Join(owner, name))
}

// GetRepoUserId gets a repo list by unique user ID.
func GetRepoUserId(c context.Context, id int64) ([]*model.Repo, error) {
	return FromContext(c).GetRepoUserId(id)
}

// GetRepoIntersect gets a repo list by account login.
func GetRepoIntersect(c context.Context, repos []*model.Repo) ([]*model.Repo, error) {
	slugs := make([]string, len(repos))
	for i, repo := range repos {
		slugs[i] = repo.Slug
	}
	return GetRepoMulti(c, slugs...)
}

// GetRepoIntersectMap gets a repo set by account login where the key is
// the repository slug and the value is the repository struct.
func GetRepoIntersectMap(c context.Context, repos []*model.Repo) (map[string]*model.Repo, error) {
	repos, err := GetRepoIntersect(c, repos)
	if err != nil {
		return nil, err
	}
	repoSet := make(map[string]*model.Repo, len(repos))
	for _, repo := range repos {
		repoSet[repo.Slug] = repo
	}
	return repoSet, nil
}

// CreateRepo creates a new repository.
func CreateRepo(c context.Context, repo *model.Repo) error {
	return FromContext(c).CreateRepo(repo)
}

// UpdateRepo updates a user repository.
func UpdateRepo(c context.Context, repo *model.Repo) error {
	return FromContext(c).UpdateRepo(repo)
}

// DeleteRepo deletes a user repository.
func DeleteRepo(c context.Context, repo *model.Repo) error {
	return FromContext(c).DeleteRepo(repo)
}

// CheckValidUser returns true if the user exists in the valid_user table
func CheckValidUser(c context.Context, login string) (bool, error) {
	return FromContext(c).CheckValidUser(login)
}

func GetValidOrgs(c context.Context) (set.Set, error) {
	//error is of no use on Get, nil coming back is what matters
	orgs, _ := cache.Get(c, "valid_orgs")
	if orgs == nil {
		orgsSlice, err := FromContext(c).GetValidOrgs()
		if err != nil {
			return nil, err
		}
		orgs = set.New(orgsSlice...)
		cache.Set(c, "valid_orgs", orgs)
	}
	return orgs.(set.Set), nil
}

func GetUserEnabledOrgs(c context.Context, names []string) ([]*model.OrgDb, error) {
	return FromContext(c).GetUserEnabledOrgs(names)
}

func GetOrgName(c context.Context, owner string) (*model.OrgDb, error) {
	return FromContext(c).GetOrgByName(owner)
}

// GetReposForOrg returns all repos in the specified org
func GetReposForOrg(c context.Context, owner string) ([]*model.Repo, error) {
	return FromContext(c).GetReposForOrg(owner)
}

// CreateOrg creates a new auto-enrolling org.
func CreateOrg(c context.Context, org *model.OrgDb) error {
	return FromContext(c).CreateOrg(org)
}

// DeleteOrg deletes an auto-enrolling org
func DeleteOrg(c context.Context, org *model.OrgDb) error {
	return FromContext(c).DeleteOrg(org)
}

func GetSlackUrl(c context.Context, hostname string, user string) (string, error) {
	return FromContext(c).GetSlackUrl(hostname, user)
}

func AddUpdateSlackUrl(c context.Context, hostname string, user string, url string) error {
	return FromContext(c).AddUpdateSlackUrl(hostname, user, url)
}

func DeleteSlackUrl(c context.Context, hostname string, user string) error {
	return FromContext(c).DeleteSlackUrl(hostname, user)
}
