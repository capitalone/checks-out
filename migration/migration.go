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
package migration

import (
	"context"

	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/store"

	"github.com/mspiegel/go-multierror"
	log "github.com/sirupsen/logrus"
)

// Migrate performs any (store x remote) operations necessary
// after a database migration.
func Migrate(r remote.Remote, s store.Store) error {
	var errs error
	ids := s.GetMigrations()
	if ids.Contains("002_org.sql") {
		errs = multierror.Append(errs, insertRepoOrgs(r, s))
	}
	if ids.Contains("005_oauth_scope.sql") {
		errs = multierror.Append(errs, insertDefaultScope(s))
	}
	return errs
}

func insertDefaultScope(s store.Store) error {
	log.Info("Applying 005_oauth_scope.sql migrations")
	count := 0
	users, err := s.GetAllUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		user.Scopes = envvars.Env.Github.Scope
		err = s.UpdateUser(user)
		if err != nil {
			log.Warnf("Unable to update user %s: %s", user.Login, err)
			continue
		}
		count++
	}
	log.Infof("Applied 005_oauth_scope.sql migrations to %d out of a total %d users", count, len(users))
	return nil
}

func insertRepoOrgs(r remote.Remote, s store.Store) error {
	log.Info("Applying 002_org.sql migrations")
	count := 0
	repos, err := s.GetAllRepos()
	if err != nil {
		return err
	}
	for _, repo := range repos {
		model, err := s.GetUser(repo.UserID)
		if err != nil {
			log.Warnf("Unable to retrieve user %s (id %d) for repo %s: %s",
				repo.Owner, repo.UserID, repo.Slug, err)
			continue
		}
		update, err := r.GetRepo(context.Background(), model, repo.Owner, repo.Name)
		if err != nil {
			log.Warnf("Unable to retrieve repo %s: %s", repo.Slug, err)
			continue
		}
		repo.Org = update.Org
		err = s.UpdateRepo(repo)
		if err != nil {
			log.Warnf("Unable to update repo %s: %s", repo.Slug, err)
			continue
		}
		count++
	}
	log.Infof("Applied 002_org.sql migrations to %d out of a total %d repos", count, len(repos))
	return nil
}
