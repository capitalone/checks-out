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
	"time"

	"github.com/capitalone/checks-out/admin"
	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/router/middleware/session"
	"github.com/capitalone/checks-out/shared/httputil"
	"github.com/capitalone/checks-out/store"

	"github.com/gin-gonic/gin"
	"github.com/jonbodner/stackerr"
	"github.com/sirupsen/logrus"
)

const (
	notAdmin   = "user is not an administrator"
	noAdminOrg = "GITHUB_ADMIN_ORG environment variable not defined"
)

func unauthorizedError() error {
	status := http.StatusUnauthorized
	return exterror.Create(status, stackerr.New(notAdmin))
}

func adminMissingError() error {
	if envvars.Env.Github.AdminOrg == "" {
		if envvars.Env.Monitor.Sunlight {
			status := http.StatusUnauthorized
			return exterror.Create(status, stackerr.New(noAdminOrg))
		}
		return unauthorizedError()
	}
	return nil
}

func CheckAdmin(c *gin.Context) {
	err := adminMissingError()
	if err != nil {
		c.Error(err)
		return
	}
	user := session.User(c)
	orgs, err := remote.GetOrgs(c, user)
	if err != nil {
		c.Error(err)
		return
	}
	isAdmin := false
	for _, o := range orgs {
		if o.Login == envvars.Env.Github.AdminOrg {
			isAdmin = true
		}
	}
	if !isAdmin {
		c.Error(unauthorizedError())
		return
	}
}

func GetAllConfigurationSubtree(c *gin.Context) {
	path := c.Param("path")
	divisor := envvars.Env.Github.RequestsHz
	if divisor < 1 {
		divisor = 1
	}
	rate := time.Second / time.Duration(divisor)
	throttle := time.Tick(rate)
	body := make(map[string]interface{})
	repos, err := store.GetAllRepos(c)
	if err != nil {
		c.Error(err)
		return
	}
	for _, r := range repos {
		<-throttle
		u, err := store.GetUser(c, r.UserID)
		if err != nil {
			body[r.Slug] = err.Error()
		} else {
			body[r.Slug] = admin.GetConfigSubtree(c, r, u, path)
		}
	}
	IndentedJSON(c, 200, body)
}

func AdminDeleteRepo(c *gin.Context) {
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
	err = store.DeleteRepo(c, repo)
	if err != nil {
		msg := fmt.Sprintf("Deleting repository %s", name)
		c.Error(exterror.Append(err, msg))
		return
	}
	link := fmt.Sprintf(
		"%s/hook",
		httputil.GetURL(c.Request),
	)

	user, err := store.GetUser(c, repo.UserID)
	if err != nil {
		msg := fmt.Sprintf("Deleting repository %s", name)
		c.Error(exterror.Append(err, msg))
		return
	}

	err = remote.DelHook(c, user, repo, link)
	if err != nil {
		ext := exterror.Convert(err)
		if ext.Status < 500 {
			logrus.Warnf("Deleting repository hook for %s. %s", name, err)
		} else {
			logrus.Errorf("Deleting repository hook for %s. %s", name, err)
		}
	}
	c.String(200, "")
}
