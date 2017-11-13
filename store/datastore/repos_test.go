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
package datastore

import (
	"testing"

	"github.com/capitalone/checks-out/model"

	"github.com/franela/goblin"
)

func Test_repostore(t *testing.T) {
	db, ids, driver := openTest()
	defer db.Close()

	s := From(db, ids, driver)
	g := goblin.Goblin(t)
	g.Describe("Repo", func() {

		// before each test be sure to purge the package
		// table data from the database.
		g.BeforeEach(func() {
			db.Exec("DELETE FROM repos")
			db.Exec("DELETE FROM users")
		})

		g.It("Should Set a Repo", func() {
			repo := model.Repo{
				UserID: 1,
				Slug:   "bradrydzewski/drone",
				Owner:  "bradrydzewski",
				Name:   "drone",
			}
			err1 := s.CreateRepo(&repo)
			err2 := s.UpdateRepo(&repo)
			getrepo, err3 := s.GetRepo(repo.ID)
			g.Assert(err1 == nil).IsTrue()
			g.Assert(err2 == nil).IsTrue()
			g.Assert(err3 == nil).IsTrue()
			g.Assert(repo.ID).Equal(getrepo.ID)
		})

		g.It("Should Add a Repo", func() {
			repo := model.Repo{
				UserID: 1,
				Slug:   "bradrydzewski/drone",
				Owner:  "bradrydzewski",
				Name:   "drone",
			}
			err := s.CreateRepo(&repo)
			g.Assert(err == nil).IsTrue()
			g.Assert(repo.ID != 0).IsTrue()
		})

		g.It("Should Get a Repo by ID", func() {
			repo := model.Repo{
				UserID:  1,
				Slug:    "bradrydzewski/drone",
				Owner:   "bradrydzewski",
				Name:    "drone",
				Link:    "https://github.com/octocat/hello-world",
				Private: true,
			}
			s.CreateRepo(&repo)
			getrepo, err := s.GetRepo(repo.ID)
			g.Assert(err == nil).IsTrue()
			g.Assert(repo.ID).Equal(getrepo.ID)
			g.Assert(repo.UserID).Equal(getrepo.UserID)
			g.Assert(repo.Owner).Equal(getrepo.Owner)
			g.Assert(repo.Name).Equal(getrepo.Name)
			g.Assert(repo.Private).Equal(getrepo.Private)
			g.Assert(repo.Link).Equal(getrepo.Link)
		})

		g.It("Should Get a Repo by Slug", func() {
			repo := model.Repo{
				UserID: 1,
				Slug:   "bradrydzewski/drone",
				Owner:  "bradrydzewski",
				Name:   "drone",
			}
			s.CreateRepo(&repo)
			getrepo, err := s.GetRepoSlug(repo.Slug)
			g.Assert(err == nil).IsTrue()
			g.Assert(repo.ID).Equal(getrepo.ID)
			g.Assert(repo.UserID).Equal(getrepo.UserID)
			g.Assert(repo.Owner).Equal(getrepo.Owner)
			g.Assert(repo.Name).Equal(getrepo.Name)
		})

		g.It("Should Get a Multiple Repos", func() {
			repo1 := &model.Repo{
				UserID: 1,
				Owner:  "foo",
				Name:   "bar",
				Slug:   "foo/bar",
			}
			repo2 := &model.Repo{
				UserID: 2,
				Owner:  "octocat",
				Name:   "fork-knife",
				Slug:   "octocat/fork-knife",
			}
			repo3 := &model.Repo{
				UserID: 2,
				Owner:  "octocat",
				Name:   "hello-world",
				Slug:   "octocat/hello-world",
			}
			s.CreateRepo(repo1)
			s.CreateRepo(repo2)
			s.CreateRepo(repo3)

			repos, err := s.GetRepoMulti("octocat/fork-knife", "octocat/hello-world")
			g.Assert(err == nil).IsTrue()
			g.Assert(len(repos)).Equal(2)
			g.Assert(repos[0].ID).Equal(repo2.ID)
			g.Assert(repos[1].ID).Equal(repo3.ID)
		})

		g.It("Should Delete a Repo", func() {
			repo := model.Repo{
				UserID: 1,
				Slug:   "bradrydzewski/drone",
				Owner:  "bradrydzewski",
				Name:   "drone",
			}
			s.CreateRepo(&repo)
			_, err1 := s.GetRepo(repo.ID)
			err2 := s.DeleteRepo(&repo)
			_, err3 := s.GetRepo(repo.ID)
			g.Assert(err1 == nil).IsTrue()
			g.Assert(err2 == nil).IsTrue()
			g.Assert(err3 == nil).IsFalse()
		})

		g.It("Should Enforce Unique Repo Name", func() {
			repo1 := model.Repo{
				UserID: 1,
				Slug:   "bradrydzewski/drone",
				Owner:  "bradrydzewski",
				Name:   "drone",
			}
			repo2 := model.Repo{
				UserID: 2,
				Slug:   "bradrydzewski/drone",
				Owner:  "bradrydzewski",
				Name:   "drone",
			}
			err1 := s.CreateRepo(&repo1)
			err2 := s.CreateRepo(&repo2)
			g.Assert(err1 == nil).IsTrue()
			g.Assert(err2 == nil).IsFalse()
		})
	})
}

func Test_orgstore(t *testing.T) {
	db, ids, driver := openTest()
	defer db.Close()

	s := From(db, ids, driver)
	g := goblin.Goblin(t)
	g.Describe("Org", func() {

		// before each test be sure to purge the package
		// table data from the database.
		g.BeforeEach(func() {
			db.Exec("DELETE FROM orgs")
			db.Exec("DELETE FROM users")
		})

		g.It("Should Add an Org and get by name", func() {
			org := model.OrgDb{
				UserID: 1,
				Owner:  "testorg",
			}
			err1 := s.CreateOrg(&org)
			getorg, err3 := s.GetOrgByName(org.Owner)
			g.Assert(err1 == nil).IsTrue()
			g.Assert(err3 == nil).IsTrue()
			g.Assert(org.ID).Equal(getorg.ID)
		})

		g.It("Should Get a Multiple Orgs", func() {
			org1 := &model.OrgDb{
				UserID: 1,
				Owner:  "foo",
			}
			org2 := &model.OrgDb{
				UserID: 2,
				Owner:  "octocat",
			}
			org3 := &model.OrgDb{
				UserID: 2,
				Owner:  "bar",
			}
			s.CreateOrg(org1)
			s.CreateOrg(org2)
			s.CreateOrg(org3)

			repos, err := s.GetUserEnabledOrgs([]string{"octocat", "bar"})
			g.Assert(err == nil).IsTrue()
			g.Assert(len(repos)).Equal(2)
			g.Assert(repos[0].Owner == "octocat" || repos[0].Owner == "bar").IsTrue("Expected to be octocat or bar")
			g.Assert(repos[1].Owner == "octocat" || repos[1].Owner == "bar").IsTrue("Expected to be octocat or bar")
		})

		g.It("Should Delete an Org", func() {
			org := model.OrgDb{
				UserID: 1,
				Owner:  "testorg",
			}
			s.CreateOrg(&org)
			_, err1 := s.GetOrgByName(org.Owner)
			err2 := s.DeleteOrg(&org)
			_, err3 := s.GetOrgByName(org.Owner)
			g.Assert(err1 == nil).IsTrue()
			g.Assert(err2 == nil).IsTrue()
			g.Assert(err3 == nil).IsFalse()
		})

		g.It("Should Enforce Unique Owners", func() {
			org1 := model.OrgDb{
				UserID: 1,
				Owner:  "foo",
			}
			org2 := model.OrgDb{
				UserID: 2,
				Owner:  "foo",
			}
			err1 := s.CreateOrg(&org1)
			err2 := s.CreateOrg(&org2)
			g.Assert(err1 == nil).IsTrue()
			g.Assert(err2 == nil).IsFalse()
		})
	})
}
