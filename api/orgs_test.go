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
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/capitalone/checks-out/cache"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/router/middleware"
	"github.com/capitalone/checks-out/store"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type mockCache struct {
	cache.Cache
}

func (mc *mockCache) Get(s string) (interface{}, error) {
	if s == "orgs:octocat" {
		return fakeOrgs, nil
	}
	return nil, errors.New("Unexpected")
}

type mockCache2 struct {
	cache.Cache
}

func (mc *mockCache2) Get(s string) (interface{}, error) {
	if s == "orgs:octocat" {
		return nil, errors.New("Not Found")
	}
	return nil, errors.New("Unexpected")
}

type mockRemote2 struct {
	remote.Remote
}

func (mr *mockRemote2) GetOrgs(c context.Context, u *model.User) ([]*model.GitHubOrg, error) {
	if u == fakeUser {
		return nil, errors.New("Not Found")
	}
	return nil, errors.New("Unexpected")
}

type mockStore struct {
	store.Store
}

func (ms *mockStore) GetUserEnabledOrgs(names []string) ([]*model.OrgDb, error) {
	return nil, nil
}
func TestOrgs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logrus.SetOutput(ioutil.Discard)

	g := goblin.Goblin(t)

	g.Describe("Org endpoint", func() {
		g.It("Should return the orgs list", func() {
			mc := &mockCache{}
			ms := &mockStore{}

			e := gin.New()
			e.NoRoute(GetOrgs)
			e.Use(func(c *gin.Context) {
				c.Set("user", fakeUser)
				c.Set("cache", mc)
				c.Set("store", ms)
			})

			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			e.ServeHTTP(w, r)

			// the user is appended to the orgs list so we retrieve a full list of
			// accounts to which the user has access.
			orgs := append(fakeOrgs, &model.GitHubOrg{
				Login: fakeUser.Login,
			})

			want, _ := json.MarshalIndent(orgs, "", "    ")
			got := strings.TrimSpace(w.Body.String())
			g.Assert(got).Equal(string(want))
			g.Assert(w.Code).Equal(200)
		})

		g.It("Should return a 500 error", func() {
			r := &mockRemote2{}
			mc := &mockCache2{}

			e := gin.New()
			e.NoRoute(GetOrgs)
			e.Use(middleware.ExtError())
			e.Use(func(c *gin.Context) {
				c.Set("user", fakeUser)
				c.Set("cache", mc)
				c.Set("remote", r)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)
			e.ServeHTTP(w, req)

			got := strings.TrimSpace(w.Body.String())
			g.Assert(got).Equal("Getting organizations for user octocat. Not Found")
			g.Assert(w.Code).Equal(500)
		})
	})
}
