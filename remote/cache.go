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
	"fmt"

	"github.com/capitalone/checks-out/cache"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/set"
)

// GetOrgs returns the list of user organizations from the cache
// associated with the current context.
func GetOrgs(c context.Context, user *model.User) ([]*model.GitHubOrg, error) {
	key := fmt.Sprintf("orgs:%s",
		user.Login,
	)
	// if we fetch from the cache we can return immediately
	val, err := cache.Get(c, key)
	if err == nil {
		return val.([]*model.GitHubOrg), nil
	}
	// else we try to grab from the remote system and
	// populate our cache.
	orgs, err := FromContext(c).GetOrgs(c, user)
	if err != nil {
		return nil, err
	}
	cache.Set(c, key, orgs)
	return orgs, nil
}

// ListTeams returns the list of team names from the cache
// associated with the current context.
func ListTeams(c context.Context, user *model.User, org string) (set.Set, error) {
	key := fmt.Sprintf("teams:%s", org)
	// if we fetch from the cache we can return immediately
	val, err := cache.Get(c, key)
	if err == nil {
		return val.(set.Set), nil
	}
	// else we try to grab from the remote system and
	// populate our cache.
	teams, err := FromContext(c).ListTeams(c, user, org)
	if err != nil {
		return nil, err
	}
	cache.Set(c, key, teams)
	return teams, nil
}

// GetPerm returns the user permissions repositories from the cache
// associated with the current repository.
func GetPerm(c context.Context, user *model.User, owner, name string) (*model.Perm, error) {
	key := fmt.Sprintf("perms:%s:%s/%s",
		user.Login,
		owner,
		name,
	)
	// if we fetch from the cache we can return immediately
	val, err := cache.Get(c, key)
	if err == nil {
		return val.(*model.Perm), nil
	}
	// else we try to grab from the remote system and
	// populate our cache.
	perm, err := FromContext(c).GetPerm(c, user, owner, name)
	if err != nil {
		return nil, err
	}
	cache.Set(c, key, perm)
	return perm, nil
}

func getPeople(c context.Context, user *model.User, ids set.Set) ([]*model.Person, error) {
	var people []*model.Person
	for login := range ids {
		key := fmt.Sprintf("people:%s", login)
		val, err := cache.Longterm.Get(key)
		if err == nil {
			people = append(people, val.(*model.Person))
		} else {
			p, err := GetPerson(c, user, login)
			if err != nil {
				return nil, err
			}
			err = cache.Longterm.Set(key, p)
			if err != nil {
				return nil, err
			}
			people = append(people, p)
		}
	}
	return people, nil
}

// GetCollaborators gets an collaborators member list from the remote system.
// Looks for information about each person in the long term cache.
func GetCollaborators(c context.Context, user *model.User, owner, name string) ([]*model.Person, error) {
	users, err := FromContext(c).GetCollaborators(c, user, owner, name)
	if err != nil {
		return nil, err
	}
	return getPeople(c, user, users)
}

// GetOrgMembers gets an organization member list from the remote system.
// Looks for information about each person in the long term cache.
func GetOrgMembers(c context.Context, user *model.User, org string) ([]*model.Person, error) {
	users, err := FromContext(c).GetOrgMembers(c, user, org)
	if err != nil {
		return nil, err
	}
	return getPeople(c, user, users)
}

// GetTeamMembers gets an repo's team members list from the remote system.
// Looks for information about each person in the long term cache.
func GetTeamMembers(c context.Context, user *model.User, org string, team string) ([]*model.Person, error) {
	users, err := FromContext(c).GetTeamMembers(c, user, org, team)
	if err != nil {
		return nil, err
	}
	return getPeople(c, user, users)
}
