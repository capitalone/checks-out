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
package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/router/middleware/session"
	"github.com/capitalone/checks-out/shared/httputil"
	"github.com/capitalone/checks-out/shared/token"
	"github.com/capitalone/checks-out/store"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// GetAllReposCount gets the count of all repositories managed by Checks-Out.
func GetAllReposCount(c *gin.Context) {
	repos, err := store.GetAllRepos(c)
	if err != nil {
		msg := "Getting repository list"
		c.Error(exterror.Append(err, msg))
		return
	}
	c.String(200, "%d", len(repos))
}

// GetAllRepos gets all public repositories managed by Checks-Out.
func GetAllRepos(c *gin.Context) {
	repos, err := store.GetAllRepos(c)
	if err != nil {
		msg := "Getting repository list"
		c.Error(exterror.Append(err, msg))
		return
	}
	public := []*model.Repo{}
	for _, r := range repos {
		if !r.Private {
			public = append(public, r)
		}
	}
	IndentedJSON(c, 200, public)
}

// GetUserRepos gets the user's repository list.
func GetUserRepos(c *gin.Context) {
	user := session.User(c)
	repos, err := remote.GetUserRepos(c, user)
	if err != nil {
		msg := "Getting remote repository list"
		c.Error(exterror.Append(err, msg))
		return
	}
	repoResponse(c, repos)
}

// GetOrgRepos gets the active org repository list for the current user.
func GetOrgRepos(c *gin.Context) {
	var (
		org = c.Param("org")
	)
	user := session.User(c)
	repos, err := remote.GetOrgRepos(c, user, org)
	if err != nil {
		msg := "Getting remote repository list"
		c.Error(exterror.Append(err, msg))
		return
	}
	repoResponse(c, repos)
}

func repoResponse(c *gin.Context, repos []*model.Repo) {
	// copy the slice since we are going to mutate it and don't
	// want any nasty data races if the slice came from the cache.
	repoc := make([]*model.Repo, len(repos))
	copy(repoc, repos)

	repom, err := store.GetRepoIntersectMap(c, repos)
	if err != nil {
		msg := "Getting active repository list"
		c.Error(exterror.Append(err, msg))
		return
	}

	// merges the slice of active and remote repositories favoring
	// and swapping in local repository information when possible.
	for i, repo := range repoc {
		r, ok := repom[repo.Slug]
		if ok {
			repoc[i] = r
		}
	}
	IndentedJSON(c, 200, repoc)
}

// GetRepo gets the repository by slug.
func GetRepo(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("repo")
	)
	repo, err := store.GetRepoOwnerName(c, owner, name)
	if err != nil {
		msg := fmt.Sprintf("Getting repository %s", name)
		c.Error(exterror.Append(err, msg))
		return
	}
	IndentedJSON(c, 200, repo)
}

// PostRepo activates a new repository.
func PostRepo(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("repo")
		user  = session.User(c)
		caps  = session.Capability(c)
	)

	baseURL := httputil.GetURL(c.Request)
	repo, err := TurnOnRepoFailFast(c, user, caps, owner, name, baseURL)
	if err != nil {
		c.Error(err)
		return
	}

	IndentedJSON(c, 200, repo)

}

// DeleteRepo deletes a repository configuration.
func DeleteRepo(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("repo")
		user  = session.User(c)
	)
	repo, err := store.GetRepoOwnerName(c, owner, name)
	if err != nil {
		msg := fmt.Sprintf("Getting repository %s", name)
		c.Error(exterror.Append(err, msg))
		return
	}
	err = TurnOffRepo(c, user, repo, owner, name, httputil.GetURL(c.Request))
	if err != nil {
		c.Error(err)
	} else {
		c.String(200, "")
	}
}

// TurnOnRepoFailFast enables the repository.
// If already activated then return an error.
// Validate the configuration files.
func TurnOnRepoFailFast(c context.Context, user *model.User, caps *model.Capabilities, owner, name, baseURL string) (*model.Repo, error) {
	return enableRepo(c, user, caps, owner, name, baseURL, false, true)
}

// TurnOnRepoQuiet enables the repository.
// If already activated then do nothing.
// Do not validate the configuration files.
func TurnOnRepoQuiet(c context.Context, user *model.User, owner, name, baseURL string) (*model.Repo, error) {
	return enableRepo(c, user, nil, owner, name, baseURL, true, false)
}

func enableRepo(c context.Context, user *model.User, caps *model.Capabilities, owner, name, baseURL string, idempotent, validate bool) (*model.Repo, error) {
	var repo *model.Repo
	var err error

	// if repo already activated
	if _, err = store.GetRepoOwnerName(c, owner, name); err == nil {
		if idempotent {
			return nil, nil
		}
		err = errors.Errorf("Unable to activate repository %s/%s because it is already active.", owner, name)
		err = exterror.Create(http.StatusConflict, err)
		return nil, err
	}

	repo, err = remote.GetRepo(c, user, owner, name)
	if err != nil {
		msg := fmt.Sprintf("Looking for repository %s on Github", name)
		return nil, exterror.Append(err, msg)
	}

	if validate {
		_, err = validateRepo(c, repo, owner, name, user, caps)
		if err != nil {
			return nil, err
		}
	}

	repo.UserID = user.ID
	repo.Secret = model.Rand()

	// creates a token to authorize the link callback url
	t := token.New(token.HookToken, repo.Slug)
	sig, err := t.Sign(repo.Secret)
	if err != nil {
		return nil, exterror.Append(err, "Activating repository")
	}

	// create the hook callback url
	link := fmt.Sprintf(
		"%s/hook?access_token=%s",
		baseURL,
		sig,
	)
	err = remote.SetHook(c, user, repo, link)
	if err != nil {
		return nil, exterror.Append(err, "Creating hook")
	}

	err = store.CreateRepo(c, repo)
	if err != nil {
		return nil, exterror.Append(err, "Activating the repository")
	}

	return repo, nil
}

func TurnOffRepo(c context.Context, user *model.User, repo *model.Repo, owner, name, baseURL string) error {

	err := store.DeleteRepo(c, repo)
	if err != nil {
		msg := fmt.Sprintf("Deleting repository %s", name)
		return exterror.Append(err, msg)
	}

	link := fmt.Sprintf(
		"%s/hook",
		baseURL,
	)
	err = remote.DelHook(c, user, repo, link)
	if err != nil {
		ext := exterror.Convert(err)
		if ext.Status < 500 {
			logrus.Warnf("Deleting repository hook for %s. %s", name, err)
		} else {
			logrus.Errorf("Deleting repository hook for %s. %s", name, err)
		}
	}

	return nil
}
