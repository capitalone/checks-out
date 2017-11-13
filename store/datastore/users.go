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

	"github.com/russross/meddler"
)

func (db *datastore) GetUser(id int64) (*model.User, error) {
	var usr = new(model.User)
	var err = meddler.Load(db, userTable, usr, id)
	return usr, err
}

func (db *datastore) GetUserLogin(login string) (*model.User, error) {
	var usr = new(model.User)
	var err = meddler.QueryRow(db, usr, userLoginQuery[db.curDB], login)
	return usr, err
}

func (db *datastore) CreateUser(user *model.User) error {
	return meddler.Insert(db, userTable, user)
}

func (db *datastore) UpdateUser(user *model.User) error {
	return meddler.Update(db, userTable, user)
}

func (db *datastore) DeleteUser(user *model.User) error {
	var _, err = db.Exec(userDeleteStmt[db.curDB], user.ID)
	return err
}

func (db *datastore) GetAllUsers() ([]*model.User, error) {
	var repos = []*model.User{}
	var err = meddler.QueryAll(db, &repos, userListQuery)
	return repos, err
}

func (db *datastore) CheckValidUser(login string) (bool, error) {
	rows, err := db.Query(limitUserQuery[db.curDB], login)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func (db *datastore) GetValidOrgs() ([]string, error) {
	out := []string{}
	rows, err := db.Query(limitOrgQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var org string
		err := rows.Scan(&org)
		if err != nil {
			return nil, err
		}
		out = append(out, org)
	}
	return out, nil
}

const userTable = "users"

var userLoginQuery = map[string]string{
	POSTGRES: `
	SELECT *
	FROM users
	WHERE user_login=$1
	LIMIT 1
	`,
	MYSQL: `
	SELECT *
	FROM users
	WHERE user_login= ?
	LIMIT 1
	`,
	SQLITE: `
	SELECT *
	FROM users
	WHERE user_login= ?
	LIMIT 1
	`,
}

const userListQuery = `
SELECT *
FROM users
`

const userCountQuery = `
SELECT count(1)
FROM users
`

var userDeleteStmt = map[string]string{
	POSTGRES: `
	DELETE FROM users
	WHERE user_id=$1
	`,
	MYSQL: `
	DELETE FROM users
	WHERE user_id=?
	`,
	SQLITE: `
	DELETE FROM users
	WHERE user_id=?
	`,
}

var limitUserQuery = map[string]string{
	POSTGRES: `
	SELECT 1 AS exist FROM limit_users
	WHERE login=$1
	`,
	MYSQL: `
	SELECT 1 AS exist FROM limit_users
	WHERE login=?
	`,
	SQLITE: `
	SELECT 1 AS exist FROM limit_users
	WHERE login=?
	`,
}

const limitOrgQuery = `
SELECT org
FROM limit_orgs
`
