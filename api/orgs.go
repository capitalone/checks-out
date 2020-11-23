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
package api

import (
	"fmt"
	"net/http"

	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/router/middleware/session"
	"github.com/capitalone/checks-out/set"
	"github.com/capitalone/checks-out/shared/httputil"
	"github.com/capitalone/checks-out/shared/token"
	"github.com/capitalone/checks-out/store"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func collectAllOrgs(c *gin.Context) ([]*model.GitHubOrg, set.Set, error) {
	user := session.User(c)
	orgs, err := remote.GetOrgs(c, user)
	if err != nil {
		msg := fmt.Sprintf("Getting organizations for user %s", user.Login)
		err = exterror.Append(err, msg)
		return nil, nil, err
	}
	var names []string
	for _, org := range orgs {
		names = append(names, org.Login)
	}
	enabled, err := store.GetUserEnabledOrgs(c, names)
	if err != nil {
		msg := fmt.Sprintf("Getting enabled organizations for user %s", user.Login)
		err = exterror.Append(err, msg)
		return nil, nil, err
	}
	enabledNames := set.New()
	for _, org := range enabled {
		enabledNames.Add(org.Owner)
	}
	return orgs, enabledNames, nil
}

// GetOrgs gets the list of user organizations.
func GetOrgs(c *gin.Context) {
	user := session.User(c)
	orgs, enabled, err := collectAllOrgs(c)
	if err != nil {
		c.Error(err)
		return
	}
	for k := range orgs {
		orgs[k].Enabled = enabled.Contains(orgs[k].Login)
	}
	orgs = append(orgs, &model.GitHubOrg{
		Login:  user.Login,
		Avatar: user.Avatar,
	})
	IndentedJSON(c, 200, orgs)
}

// GetEnabledOrgs gets the list of organizations monitored by the service.
func GetEnabledOrgs(c *gin.Context) {
	_, enabled, err := collectAllOrgs(c)
	if err != nil {
		c.Error(err)
		return
	}
	orgs := []*model.GitHubOrg{}
	for name := range enabled {
		orgs = append(orgs, &model.GitHubOrg{
			Login:   name,
			Enabled: true,
		})
	}
	IndentedJSON(c, 200, orgs)
}

// GetTeams gets the list of teams.
func GetTeams(c *gin.Context) {
	user := session.User(c)
	owner := c.Param("owner")
	teams, err := remote.ListTeams(c, user, owner)
	if err != nil {
		msg := fmt.Sprintf("Getting organizations for user %s", user.Login)
		c.Error(exterror.Append(err, msg))
		return
	}
	IndentedJSON(c, 200, teams)
}

// PostRepo activates a new repository.
func PostOrg(c *gin.Context) {
	var (
		owner = c.Param("owner")
		user  = session.User(c)
	)

	// verify repo doesn't already exist
	if _, err := store.GetOrgName(c, owner); err == nil {
		err = errors.Errorf("Unable to activate org %s because it is already active.", owner)
		inner := exterror.Create(http.StatusConflict, err)
		c.Error(inner)
		return
	}

	org, err := remote.GetOrg(c, user, owner)
	if err != nil {
		msg := fmt.Sprintf("Looking for org %s on Github", owner)
		c.Error(exterror.Append(err, msg))
		return
	}

	org.UserID = user.ID
	org.Secret = model.Rand()

	// creates a token to authorize the link callback url
	t := token.New(token.HookToken, org.Owner)
	sig, err := t.Sign(org.Secret)
	if err != nil {
		c.Error(exterror.Append(err, "Activating Org"))
		return
	}

	// create the hook callback url
	link := fmt.Sprintf(
		"%s/hook?access_token=%s",
		httputil.GetURL(c.Request),
		sig,
	)
	err = remote.SetOrgHook(c, user, org, link)
	if err != nil {
		c.Error(exterror.Append(err, "Creating org hook"))
		return
	}

	err = store.CreateOrg(c, org)
	if err != nil {
		c.Error(exterror.Append(err, "Activating the Org"))
		return
	}

	repos, err := remote.GetOrgRepos(c, user, owner)
	if err != nil {
		c.Error(exterror.Append(err, "Getting Org Repos"))
		return
	}
	baseURL := httputil.GetURL(c.Request)
	for _, v := range repos {
		_, err := TurnOnRepoQuiet(c, user, owner, v.Name, baseURL)
		if err != nil {
			c.Error(exterror.Append(err, "turning on Repo "+v.Name))
			return
		}
	}
	IndentedJSON(c, 200, org)
}

// AdminDeleteOrg deletes any org and all repos for that org from the database and unregisters their hooks.
// only should be called by an admin user, as it works across accounts.
func AdminDeleteOrg(c *gin.Context) {
	var (
		owner = c.Param("owner")
	)
	org, err := store.GetOrgName(c, owner)
	// verify repo already exists
	if err != nil {
		err = errors.Errorf("Unable to deactivate org %s because it is not active.", owner)
		inner := exterror.Create(http.StatusConflict, err)
		c.Error(inner)
		return
	}
	user, err := store.GetUser(c, org.UserID)
	if err != nil {
		msg := fmt.Sprintf("Deleting org %s", owner)
		c.Error(exterror.Append(err, msg))
		return
	}
	deleteOrgInner(c, owner, user, org)
}

// DeleteOrg deletes an org and all repos for that org from the database and unregisters their hooks.
func DeleteOrg(c *gin.Context) {
	var (
		owner = c.Param("owner")
		user  = session.User(c)
	)
	org, err := store.GetOrgName(c, owner)
	// verify repo already exists
	if err != nil {
		err = errors.Errorf("Unable to deactivate org %s because it is not active.", owner)
		inner := exterror.Create(http.StatusConflict, err)
		c.Error(inner)
		return
	}
	deleteOrgInner(c, owner, user, org)
}

func deleteOrgInner(c *gin.Context, owner string, user *model.User, org *model.OrgDb) {
	err := store.DeleteOrg(c, org)
	if err != nil {
		msg := fmt.Sprintf("Deleting org %s", owner)
		c.Error(exterror.Append(err, msg))
		return
	}
	link := fmt.Sprintf(
		"%s/hook",
		httputil.GetURL(c.Request),
	)
	err = remote.DelOrgHook(c, user, org, link)
	if err != nil {
		ext := exterror.Convert(err)
		if ext.Status < 500 {
			logrus.Warnf("Deleting org hook for %s. %s", owner, err)
		} else {
			logrus.Errorf("Deleting org hook for %s. %s", owner, err)
		}
	}

	repos, err := store.GetReposForOrg(c, owner)
	if err != nil {
		c.Error(exterror.Append(err, "Getting Org Repos"))
		return
	}
	for _, v := range repos {
		repo, err := store.GetRepoOwnerName(c, owner, v.Name)
		if err != nil {
			msg := fmt.Sprintf("Getting repository %s", v.Name)
			c.Error(exterror.Append(err, msg))
			return
		}
		err = TurnOffRepo(c, user, repo, owner, v.Name, httputil.GetURL(c.Request))
		if err != nil {
			c.Error(exterror.Append(err, "turning off Repo "+v.Name))
			return
		}
	}
	c.String(200, "")
}
