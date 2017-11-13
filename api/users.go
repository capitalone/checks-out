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
	"database/sql"
	"net/http"

	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/router/middleware/session"
	"github.com/capitalone/checks-out/store"

	"github.com/gin-gonic/gin"
)

// GetUser gets the currently authenticated user.
func GetUser(c *gin.Context) {
	IndentedJSON(c, 200, session.User(c))
}

// DeleteUser removes the currently authenticated user
// and all associated repositories from the database.
func DeleteUser(c *gin.Context) {
	user := session.User(c)
	repos, err := store.GetRepoUserId(c, user.ID)
	if err != nil {
		c.Error(exterror.Append(err, "Deleting user"))
		return
	}
	for _, repo := range repos {
		err = store.DeleteRepo(c, repo)
		if err != nil {
			c.Error(exterror.Append(err, "Deleting user"))
			return
		}
	}
	err = store.DeleteUser(c, user)
	if err != nil {
		c.Error(exterror.Append(err, "Deleting user"))
		return
	}
	err = remote.RevokeAuthorization(c, user)
	if err != nil {
		c.Error(exterror.Append(err, "Deleting user"))
		return
	}
	c.String(204, "")
}

func GetReposForUserLogin(c *gin.Context) {
	var (
		login = c.Param("user")
	)
	user, err := store.GetUserLogin(c, login)
	if err != nil {
		if err == sql.ErrNoRows {
			err = exterror.Create(http.StatusNotFound, err)
		}
		c.Error(err)
		return
	}
	repos, err := store.GetRepoUserId(c, user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = exterror.Create(http.StatusNotFound, err)
		}
		c.Error(err)
		return
	}
	IndentedJSON(c, 200, repos)
}
