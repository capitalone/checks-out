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
	"github.com/capitalone/checks-out/model"
	"github.com/franela/goblin"
	"testing"
)

func TestDatastore_GetValidOrgs(t *testing.T) {
	db, ids, driver := openTest()
	defer db.Close()
	s := From(db, ids, driver)

	//put in some sample values
	db.Exec("insert into limit_orgs values ('beatles'), ('stones'), ('ledzep')")
	vals, err := s.GetValidOrgs()
	if err != nil {
		t.Errorf("Error getting valid orgs: %v", err)
	}
	if len(vals) != 3 {
		t.Errorf("Expected 3 values, got %d", len(vals))
	}
	db.Exec("delete from limit_orgs")
	vals, err = s.GetValidOrgs()
	if err != nil {
		t.Errorf("Error getting valid orgs: %v", err)
	}
	if len(vals) != 0 {
		t.Errorf("Expected 0 values, got %d", len(vals))
	}
}

func TestDatastore_CheckUserValid(t *testing.T) {
	db, ids, driver := openTest()
	defer db.Close()
	s := From(db, ids, driver)

	//put in some sample values
	db.Exec("insert into limit_users values ('john'), ('paul'), ('george'), ('ringo')")
	found, err := s.CheckValidUser("john")
	if err != nil {
		t.Errorf("Error checking valid user: %v", err)
	}
	if !found {
		t.Error("Expected john to be found, he wasn't.")
	}
	found, err = s.CheckValidUser("frank")
	if err != nil {
		t.Errorf("Error checking valid user: %v", err)
	}
	if found {
		t.Error("Expected frank to not be found, he was.")
	}
	db.Exec("delete from limit_users")
	found, err = s.CheckValidUser("john")
	if err != nil {
		t.Errorf("Error checking valid user: %v", err)
	}
	if found {
		t.Error("Expected john to not be found, he was.")
	}
	found, err = s.CheckValidUser("frank")
	if err != nil {
		t.Errorf("Error checking valid user: %v", err)
	}
	if found {
		t.Error("Expected frank to not be found, he was.")
	}
}

func Test_userstore(t *testing.T) {
	db, ids, driver := openTest()
	defer db.Close()
	s := From(db, ids, driver)

	g := goblin.Goblin(t)
	g.Describe("User", func() {

		// before each test be sure to purge the package
		// table data from the database.
		g.BeforeEach(func() {
			db.Exec("DELETE FROM users")
		})

		g.It("Should Update a User", func() {
			user := model.User{
				Login: "joe",
				Token: "e42080dddf012c718e476da161d21ad5",
			}
			err1 := s.CreateUser(&user)
			err2 := s.UpdateUser(&user)
			getuser, err3 := s.GetUser(user.ID)
			g.Assert(err1 == nil).IsTrue()
			g.Assert(err2 == nil).IsTrue()
			g.Assert(err3 == nil).IsTrue()
			g.Assert(user.ID).Equal(getuser.ID)
		})

		g.It("Should Add a new User", func() {
			user := model.User{
				Login: "joe",
				Token: "e42080dddf012c718e476da161d21ad5",
			}
			err := s.CreateUser(&user)
			g.Assert(err == nil).IsTrue()
			g.Assert(user.ID != 0).IsTrue()
		})

		g.It("Should Get a User", func() {
			user := model.User{
				Login:  "joe",
				Token:  "f0b461ca586c27872b43a0685cbc2847",
				Secret: "976f22a5eef7caacb7e678d6c52f49b1",
				Avatar: "b9015b0857e16ac4d94a0ffd9a0b79c8",
			}

			s.CreateUser(&user)
			getuser, err := s.GetUser(user.ID)
			g.Assert(err == nil).IsTrue()
			g.Assert(user.ID).Equal(getuser.ID)
			g.Assert(user.Login).Equal(getuser.Login)
			g.Assert(user.Token).Equal(getuser.Token)
			g.Assert(user.Secret).Equal(getuser.Secret)
			g.Assert(user.Avatar).Equal(getuser.Avatar)
		})

		g.It("Should Get a User By Login", func() {
			user := model.User{
				Login: "joe",
				Token: "e42080dddf012c718e476da161d21ad5",
			}
			s.CreateUser(&user)
			getuser, err := s.GetUserLogin(user.Login)
			g.Assert(err == nil).IsTrue()
			g.Assert(user.ID).Equal(getuser.ID)
			g.Assert(user.Login).Equal(getuser.Login)
		})

		g.It("Should Enforce Unique User Login", func() {
			user1 := model.User{
				Login: "joe",
				Token: "e42080dddf012c718e476da161d21ad5",
			}
			user2 := model.User{
				Login: "joe",
				Token: "ab20g0ddaf012c744e136da16aa21ad9",
			}
			err1 := s.CreateUser(&user1)
			err2 := s.CreateUser(&user2)
			g.Assert(err1 == nil).IsTrue()
			g.Assert(err2 == nil).IsFalse()
		})

		g.It("Should Del a User", func() {
			user := model.User{
				Login: "joe",
				Token: "e42080dddf012c718e476da161d21ad5",
			}
			s.CreateUser(&user)
			_, err1 := s.GetUser(user.ID)
			err2 := s.DeleteUser(&user)
			_, err3 := s.GetUser(user.ID)
			g.Assert(err1 == nil).IsTrue()
			g.Assert(err2 == nil).IsTrue()
			g.Assert(err3 == nil).IsFalse()
		})
	})
}
