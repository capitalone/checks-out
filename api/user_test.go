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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/capitalone/checks-out/model"

	"github.com/Sirupsen/logrus"
	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"
)

func TestUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logrus.SetOutput(ioutil.Discard)

	g := goblin.Goblin(t)

	g.Describe("User endpoint", func() {
		g.It("Should return the authenticated user", func() {

			e := gin.New()
			e.NoRoute(GetUser)
			e.Use(func(c *gin.Context) {
				c.Set("user", fakeUser)
			})

			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			e.ServeHTTP(w, r)

			want, _ := json.MarshalIndent(fakeUser, "", "    ")
			got := strings.TrimSpace(w.Body.String())
			g.Assert(got).Equal(string(want))
			g.Assert(w.Code).Equal(200)
		})
	})
}

var (
	fakeUser = &model.User{Login: "octocat"}
	fakeOrgs = []*model.GitHubOrg{
		{Login: "drone"},
		{Login: "docker"},
	}
)
