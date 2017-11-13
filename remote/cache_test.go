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
package remote

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/capitalone/checks-out/cache"
	"github.com/capitalone/checks-out/model"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

type MockRemote struct {
	Remote
	mock.Mock
}

func (mr *MockRemote) GetPerm(c context.Context, u *model.User, owner string, name string) (*model.Perm, error) {
	args := mr.Called(u, owner, name)
	perm, _ := args.Get(0).(*model.Perm)
	return perm, args.Error(1)
}

func (mr *MockRemote) GetRepos(c context.Context, u *model.User) ([]*model.Repo, error) {
	args := mr.Called(u)
	repos, _ := args.Get(0).([]*model.Repo)
	return repos, args.Error(1)
}

func (mr *MockRemote) GetOrgs(c context.Context, u *model.User) ([]*model.GitHubOrg, error) {
	args := mr.Called(u)
	orgs, _ := args.Get(0).([]*model.GitHubOrg)
	return orgs, args.Error(1)
}

func TestHelper(t *testing.T) {

	g := goblin.Goblin(t)

	g.Describe("Cache helpers", func() {

		var c *gin.Context
		var r *MockRemote

		g.BeforeEach(func() {
			c = new(gin.Context)
			cache.ToContext(c, cache.Default())

			r = &MockRemote{}
			ToContext(c, r)
		})

		g.It("Should get permissions from remote", func() {
			r.On("GetPerm", fakeUser, fakeRepo.Owner, fakeRepo.Name).Return(fakePerm, nil).Once()
			p, err := GetPerm(c, fakeUser, fakeRepo.Owner, fakeRepo.Name)
			g.Assert(p).Equal(fakePerm)
			g.Assert(err).Equal(nil)
		})

		g.It("Should get permissions from cache", func() {
			key := fmt.Sprintf("perms:%s:%s/%s",
				fakeUser.Login,
				fakeRepo.Owner,
				fakeRepo.Name,
			)

			cache.Set(c, key, fakePerm)
			r.On("GetPerm", fakeUser, fakeRepo.Owner, fakeRepo.Name).Return(nil, fakeErr).Once()
			p, err := GetPerm(c, fakeUser, fakeRepo.Owner, fakeRepo.Name)
			g.Assert(p).Equal(fakePerm)
			g.Assert(err).Equal(nil)
		})

		g.It("Should get permissions error", func() {
			r.On("GetPerm", fakeUser, fakeRepo.Owner, fakeRepo.Name).Return(nil, fakeErr).Once()
			p, err := GetPerm(c, fakeUser, fakeRepo.Owner, fakeRepo.Name)
			g.Assert(p == nil).IsTrue()
			g.Assert(err).Equal(fakeErr)
		})

		g.It("Should set and get orgs", func() {
			r.On("GetOrgs", fakeUser).Return(fakeOrgs, nil).Once()
			p, err := GetOrgs(c, fakeUser)
			g.Assert(p).Equal(fakeOrgs)
			g.Assert(err).Equal(nil)
		})

		g.It("Should get orgs", func() {
			key := fmt.Sprintf("orgs:%s",
				fakeUser.Login,
			)

			cache.Set(c, key, fakeOrgs)
			r.On("GetOrgs", fakeUser).Return(nil, fakeErr).Once()
			p, err := GetOrgs(c, fakeUser)
			g.Assert(p).Equal(fakeOrgs)
			g.Assert(err).Equal(nil)
		})

		g.It("Should get org error", func() {
			r.On("GetOrgs", fakeUser).Return(nil, fakeErr).Once()
			p, err := GetOrgs(c, fakeUser)
			g.Assert(p == nil).IsTrue()
			g.Assert(err).Equal(fakeErr)
		})

	})
}

var (
	fakeErr   = errors.New("Not Found")
	fakeUser  = &model.User{Login: "octocat"}
	fakePerm  = &model.Perm{Pull: true, Push: true, Admin: true}
	fakeRepo  = &model.Repo{Owner: "octocat", Name: "Hello-World"}
	fakeRepos = []*model.Repo{
		{Owner: "octocat", Name: "Hello-World"},
		{Owner: "octocat", Name: "hello-world"},
		{Owner: "octocat", Name: "Spoon-Knife"},
	}
	fakeOrgs = []*model.GitHubOrg{
		{Login: "drone"},
		{Login: "docker"},
	}
)
