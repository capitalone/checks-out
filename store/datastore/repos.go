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
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/set"

	"github.com/russross/meddler"
)

func (db *datastore) GetMigrations() set.Set {
	return db.Migrations
}

func (db *datastore) GetRepo(id int64) (*model.Repo, error) {
	var repo = new(model.Repo)
	var err = meddler.Load(db, repoTable, repo, id)
	return repo, err
}

func (db *datastore) GetRepoSlug(slug string) (*model.Repo, error) {
	var repo = new(model.Repo)
	var err = meddler.QueryRow(db, repo, repoSlugQuery[db.curDB], slug)
	if err == sql.ErrNoRows {
		return repo, exterror.Create(http.StatusNotFound, err)
	}
	return repo, err
}

func (db *datastore) GetRepoMulti(slug ...string) ([]*model.Repo, error) {
	var repos = []*model.Repo{}
	if len(slug) == 0 {
		return repos, nil
	}
	var instr, params = db.toList(slug)
	var stmt = fmt.Sprintf(repoSlugsQuery, instr)
	var err = meddler.QueryAll(db, &repos, stmt, params...)
	return repos, err
}

func (db *datastore) GetAllRepos() ([]*model.Repo, error) {
	var repos = []*model.Repo{}
	var err = meddler.QueryAll(db, &repos, repoListQuery)
	return repos, err
}

func (db *datastore) GetUserEnabledOrgs(names []string) ([]*model.OrgDb, error) {
	var orgs = []*model.OrgDb{}
	var instr, params = db.toList(names)
	var stmt = fmt.Sprintf(userOrgListQuery, instr)
	var err = meddler.QueryAll(db, &orgs, stmt, params...)
	return orgs, err
}

func (db *datastore) GetOrgByName(owner string) (*model.OrgDb, error) {
	var org = new(model.OrgDb)
	var err = meddler.QueryRow(db, org, orgNameQuery[db.curDB], owner)
	if err == sql.ErrNoRows {
		return org, exterror.Create(http.StatusNotFound, err)
	}
	return org, err
}

// CreateOrg creates a new auto-enrolling org.
func (db *datastore) CreateOrg(org *model.OrgDb) error {
	return meddler.Insert(db, orgTable, org)
}

// DeleteOrg deletes an auto-enrolling org.
func (db *datastore) DeleteOrg(org *model.OrgDb) error {
	var _, err = db.Exec(orgDeleteStmt[db.curDB], org.ID)
	return err
}

func (db *datastore) GetReposForOrg(owner string) ([]*model.Repo, error) {
	var repos = []*model.Repo{}
	var err = meddler.QueryAll(db, &repos, repoOwnerQuery[db.curDB], owner)
	return repos, err
}

func (db *datastore) GetRepoUserId(id int64) ([]*model.Repo, error) {
	var repos = []*model.Repo{}
	var err = meddler.QueryAll(db, &repos, repoUserIdQuery[db.curDB], id)
	return repos, err
}

func (db *datastore) CreateRepo(repo *model.Repo) error {
	return meddler.Insert(db, repoTable, repo)
}

func (db *datastore) UpdateRepo(repo *model.Repo) error {
	return meddler.Update(db, repoTable, repo)
}

func (db *datastore) DeleteRepo(repo *model.Repo) error {
	var _, err = db.Exec(repoDeleteStmt[db.curDB], repo.ID)
	return err
}

func (db *datastore) toList(items []string) (string, []interface{}) {
	var size = len(items)
	if size > 990 {
		size = 990
		items = items[:990]
	}
	var qs []string
	var out []interface{}
	for i, item := range items {
		switch db.curDB {
		case POSTGRES:
			qs = append(qs, fmt.Sprintf("$%d", i+1))
		default:
			qs = append(qs, "?")
		}
		out = append(out, item)
	}
	return strings.Join(qs, ","), out
}

const (
	repoTable = "repos"
	orgTable  = "orgs"
)

var repoSlugQuery = map[string]string{
	POSTGRES: `
	SELECT *
	FROM repos
	WHERE repo_slug = $1
	LIMIT 1;
	`,
	MYSQL: `
	SELECT *
	FROM repos
	WHERE repo_slug = ?
	LIMIT 1;
	`,
	SQLITE: `
	SELECT *
	FROM repos
	WHERE repo_slug = ?
	LIMIT 1;
	`,
}

const repoSlugsQuery = `
SELECT *
FROM repos
WHERE repo_slug IN (%s)
ORDER BY repo_slug
`

var repoUserIdQuery = map[string]string{
	POSTGRES: `
	SELECT *
	FROM repos
	WHERE repo_user_id = $1
	`,
	MYSQL: `
	SELECT *
	FROM repos
	WHERE repo_user_id = ?
	`,
	SQLITE: `
	SELECT *
	FROM repos
	WHERE repo_user_id = ?
	`,
}

var repoOwnerQuery = map[string]string{
	POSTGRES: `
	SELECT *
	FROM repos
	WHERE repo_owner = $1
	`,
	MYSQL: `
	SELECT *
	FROM repos
	WHERE repo_owner = ?
	`,
	SQLITE: `
	SELECT *
	FROM repos
	WHERE repo_owner = ?
	`,
}

const repoListQuery = `
SELECT *
FROM repos
`

var orgNameQuery = map[string]string{
	POSTGRES: `
	SELECT *
	FROM orgs
	WHERE org_owner = $1
	LIMIT 1;
	`,
	MYSQL: `
	SELECT *
	FROM orgs
	WHERE org_owner = ?
	LIMIT 1;
	`,
	SQLITE: `
	SELECT *
	FROM orgs
	WHERE org_owner = ?
	LIMIT 1;
	`,
}

const userOrgListQuery = `
SELECT *
FROM orgs
WHERE org_owner IN (%s)
`

var repoDeleteStmt = map[string]string{
	POSTGRES: `
	DELETE FROM repos
	WHERE repo_id = $1
	`,
	MYSQL: `
	DELETE FROM repos
	WHERE repo_id = ?
	`,
	SQLITE: `
	DELETE FROM repos
	WHERE repo_id = ?
	`,
}

var orgDeleteStmt = map[string]string{
	POSTGRES: `
	DELETE FROM orgs
	WHERE org_id = $1
	`,
	MYSQL: `
	DELETE FROM orgs
	WHERE org_id = ?
	`,
	SQLITE: `
	DELETE FROM orgs
	WHERE org_id = ?
	`,
}
