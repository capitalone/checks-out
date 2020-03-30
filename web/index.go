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
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/router/middleware/session"
	"github.com/capitalone/checks-out/shared/token"

	"github.com/capitalone/checks-out/envvars"
	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	user := session.User(c)

	docsUrl := envvars.Env.Monitor.DocsUrl

	switch {
	case user == nil:
		c.HTML(200, "brand.html", gin.H{"DocsUrl": docsUrl})
	default:
		teams, _ := remote.GetOrgs(c, user)
		csrf, _ := token.New(token.CsrfToken, user.Login).Sign(user.Secret)
		c.HTML(200, "index.html", gin.H{"user": user, "csrf": csrf, "teams": teams, "docsUrl": docsUrl})
	}
}
