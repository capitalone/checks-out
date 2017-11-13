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
package session

import (
	"net/http"

	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/shared/token"
	"github.com/capitalone/checks-out/store"

	"github.com/gin-gonic/gin"
)

func User(c *gin.Context) *model.User {
	v, ok := c.Get("user")
	if !ok {
		return nil
	}
	u, ok := v.(*model.User)
	if !ok {
		return nil
	}
	return u
}

func UserMust(c *gin.Context) {
	user := User(c)
	switch {
	case user == nil:
		c.String(http.StatusUnauthorized,
			"You must be logged in and authorized to use this endpoint")
		c.Abort()
	default:
		c.Next()
	}
}

func SetUser(c *gin.Context) {
	var user *model.User

	// authenticates the user via an authentication cookie
	// or an auth token.
	t, err := token.ParseRequest(c.Request, func(t *token.Token) (string, error) {
		var err error
		user, err = store.GetUserLogin(c, t.Text)
		return user.Secret, err
	})

	if err == nil {
		c.Set("user", user)

		// if this is a session token (ie not the API token)
		// this means the user is accessing with a web browser,
		// so we should implement CSRF protection measures.
		if t.Kind == token.SessToken {
			err = token.CheckCsrf(c.Request, func(t *token.Token) (string, error) {
				return user.Secret, nil
			})
			// if csrf token validation fails, exit immediately
			// with a not authorized error.
			if err != nil {
				c.String(http.StatusUnauthorized,
					"You must be logged in and authorized to use this endpoint")
				c.Abort()
				return
			}
		}
	}
	c.Next()
}
