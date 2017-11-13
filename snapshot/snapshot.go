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
package snapshot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/set"
	"github.com/capitalone/checks-out/store"

	"github.com/mspiegel/go-multierror"
)

var ConfigTemplateName = fmt.Sprintf("template.%s", envvars.Env.Branding.Name)

var configFileName = fmt.Sprintf(".%s", envvars.Env.Branding.Name)

var orgRepoName = fmt.Sprintf("%s-configuration", envvars.Env.Branding.Name)

const MaintainersTemplateName = "template.MAINTAINERS"

type OrgLazyLoad struct {
	Input  set.Set             `json:"input"`
	People set.Set             `json:"people"`
	Err    error               `json:"error"`
	Init   bool                `json:"-"`
	Ctx    context.Context     `json:"-"`
	User   *model.User         `json:"-"`
	Caps   *model.Capabilities `json:"-"`
	Repo   *model.Repo         `json:"-"`
}

func findConfig(c context.Context, user *model.User, caps *model.Capabilities, repo *model.Repo) (*model.Config, error) {
	var rcfile []byte
	var exterr, err error
	// look for configuration file in current repository
	rcfile, exterr = remote.GetContents(c, user, repo, configFileName)
	if exterr == nil {
		return model.ParseConfig(rcfile, caps)
	}
	// look for legacy file
	rcfile, err = remote.GetContents(c, user, repo, ".lgtm")
	if err == nil {
		return model.ParseOldConfig(rcfile)
	}
	// look for template configuration file in org repository
	if !repo.Org || len(orgRepoName) == 0 {
		return nil, exterr
	}
	orgRepo := *repo
	orgRepo.Name = orgRepoName
	_, err = remote.GetRepo(c, user, orgRepo.Owner, orgRepo.Name)
	if err != nil {
		ext, ok := err.(exterror.ExtError)
		if ok && ext.Status == http.StatusNotFound {
			return nil, exterr
		}
		return nil, multierror.Append(exterr, err)
	}
	rcfile, err = remote.GetContents(c, user, &orgRepo, ConfigTemplateName)
	if err == nil {
		return model.ParseConfig(rcfile, caps)
	}
	return nil, multierror.Append(exterr, err)
}

func GetConfig(c context.Context, user *model.User, caps *model.Capabilities, repo *model.Repo) (*model.Config, error) {
	config, err := findConfig(c, user, caps, repo)
	if err != nil {
		err = badRequest(err)
		return nil, exterror.Append(err, fmt.Sprintf("Parsing %s file", configFileName))
	}
	if config.Deployment.Enable {
		var deployFile []byte
		deployFile, err = remote.GetContents(c, user, repo, config.Deployment.Path)
		if err != nil {
			msg := fmt.Sprintf("%s file not found", config.Deployment.Path)
			return nil, exterror.Append(err, msg)
		}
		config.LoadDeploymentMap(deployFile)
	}
	return config, nil
}

func GetConfigAndMaintainers(c context.Context, user *model.User, caps *model.Capabilities, repo *model.Repo) (*model.Config, *model.MaintainerSnapshot, error) {
	config, err := GetConfig(c, user, caps, repo)
	if err != nil {
		return nil, nil, err
	}
	snapshot, err := createSnapshot(c, user, caps, repo, config)
	if err != nil {
		return nil, nil, err
	}
	return config, snapshot, err
}

func FixSlackTargets(c context.Context, config *model.Config, user string) error {
	for k, v := range config.Comment.Targets {
		// we're potentially going to modify the curTarget
		curTarget := &config.Comment.Targets[k]
		//skip over the github targets
		if v.Target == model.Github.String() {
			continue
		}
		if v.Target == model.Slack.String() {
			//set URL to the default slack URL if no override URL was specified
			if v.Url == "" {
				curTarget.Url = envvars.Env.Slack.TargetUrl
			}
			continue
		}
		// for everything else try to find the url for it
		// if found, change target to slack and set the url
		url, err := store.GetSlackUrl(c, v.Target, user)
		if err != nil {
			return err
		}
		if url == "" {
			url, err = store.GetSlackUrl(c, v.Target, "")
		}
		if err != nil {
			return err
		}
		if url == "" {
			return errors.New("No URL found for slack host " + v.Target)
		}
		curTarget.Target = model.Slack.String()
		curTarget.Url = url
	}
	return nil
}

func findMaintainers(c context.Context, user *model.User, repo *model.Repo, path string) ([]byte, error) {
	file, err := remote.GetContents(c, user, repo, path)
	if err == nil {
		return file, nil
	}
	if !repo.Org || len(orgRepoName) == 0 {
		return nil, exterror.Append(err, fmt.Sprintf("%s file not found", path))
	}
	orgRepo := *repo
	orgRepo.Name = orgRepoName
	file, err = remote.GetContents(c, user, &orgRepo, MaintainersTemplateName)
	if err == nil {
		return file, nil
	}
	return nil, exterror.Append(err, fmt.Sprintf("%s file not found", path))
}

func createSnapshot(c context.Context, user *model.User, caps *model.Capabilities, repo *model.Repo, config *model.Config) (*model.MaintainerSnapshot, error) {
	file, err := findMaintainers(c, user, repo, config.Maintainers.Path)
	if err != nil {
		return nil, err
	}
	maintainer, err := ParseMaintainer(c, user, file, repo, config.Maintainers.Type)
	if err != nil {
		msg := fmt.Sprintf("Parsing maintainers file with %s format",
			config.Maintainers.Type)
		return nil, exterror.Append(err, msg)
	}
	snapshot, err := maintainerToSnapshot(c, user, caps, repo, maintainer)
	if err != nil {
		return nil, err
	}
	err = validateSnapshot(config, snapshot)
	return snapshot, err
}

// orgLazyLoadCandidate returns true if this github team is eligible
// for lazy expansion. The special name "_" is introduced
// when the "github-team repo-self" is found in the MAINTAINERS
// file. If "github-team repo-self" is not found then we must use eager
// evaluation of the GitHub teams in order to populate the
// model.MaintainerSnapshot.People field.
func orgLazyLoadCandidate(people set.Set, orgs map[string]*model.OrgSerde) bool {
	if _, ok := orgs["_"]; !ok {
		return false
	}
	if len(people) != 1 {
		return false
	}
	team, _ := ParseTeamName(people.Keys()[0])
	return len(team) > 0
}

func maintainerToSnapshot(c context.Context, u *model.User, caps *model.Capabilities, r *model.Repo, m *model.Maintainer) (*model.MaintainerSnapshot, error) {
	var errs error
	s := new(model.MaintainerSnapshot)
	s.People = map[string]*model.Person{}
	s.Org = map[string]model.Org{}
	for k, v := range m.RawPeople {
		k = strings.ToLower(k)
		s.People[k] = v
	}
	for k, v := range m.RawOrg {
		k = strings.ToLower(k)
		if s.People[k] != nil {
			msg := fmt.Errorf("%s cannot be both a team and a person", k)
			errs = multierror.Append(errs, badRequest(msg))
		}
		if orgLazyLoadCandidate(v.People, m.RawOrg) {
			s.Org[k] = &OrgLazyLoad{
				Input: v.People,
				Init:  false,
				Ctx:   c,
				User:  u,
				Caps:  caps,
				Repo:  r,
			}
		} else {
			org := set.Empty()
			s.Org[k] = &model.OrgSerde{
				People: org,
			}
			for login := range v.People {
				lst, err := memberExpansion(c, u, caps, r, login)
				if err != nil {
					msg := fmt.Sprintf("Attempting to expand %s", login)
					errs = multierror.Append(errs, exterror.Append(err, msg))
				} else {
					if lst != nil {
						addToPeople(lst, s)
						addToOrg(lst, org)
					} else {
						org.Add(strings.ToLower(login))
					}
				}
			}
		}
	}
	if errs != nil {
		return nil, errs
	}
	return s, nil
}

func memberExpansion(c context.Context, u *model.User, caps *model.Capabilities, r *model.Repo, input string) ([]*model.Person, error) {

	org := ParseOrgName(input)
	if org != "" {
		if org == SelfRepo {
			if !r.Org {
				p, err := remote.GetPerson(c, u, u.Login)
				if err != nil {
					return nil, exterror.Append(err, "Cannot fetch information about repository owner.")
				}
				return []*model.Person{p}, nil
			}
			if !caps.Org.Read {
				err := errors.New("Cannot read GitHub organizations with provided OAuth scopes")
				return nil, badRequest(err)
			}
			return remote.GetOrgMembers(c, u, r.Owner)
		}
		return remote.GetOrgMembers(c, u, org)
	}
	collab := ParseCollabName(input)
	if collab != "" {
		if collab == SelfRepo {
			return remote.GetCollaborators(c, u, r.Owner, r.Name)
		}
		pieces := strings.Split(collab, "/")
		if len(pieces) != 2 {
			err := fmt.Errorf("%s is not a repository slug", collab)
			return nil, badRequest(err)
		}
		return remote.GetCollaborators(c, u, pieces[0], pieces[1])
	}
	team, org := ParseTeamName(input)
	if team != "" {
		if org == "" && !r.Org {
			err := fmt.Errorf("Cannot expand GitHub teams for user repository %s/%s",
				r.Owner, r.Name)
			return nil, badRequest(err)
		}
		if !caps.Org.Read {
			err := errors.New("Cannot read GitHub teams with provided OAuth scopes")
			return nil, badRequest(err)
		}
		if org == "" {
			org = r.Owner
		}
		return remote.GetTeamMembers(c, u, org, team)
	}
	return nil, nil
}

func addToPeople(lst []*model.Person, ms *model.MaintainerSnapshot) {
	for _, m := range lst {
		login := strings.ToLower(m.Login)
		if _, ok := ms.People[login]; !ok {
			ms.People[login] = m
		}
	}
}

func addToOrg(lst []*model.Person, org set.Set) {
	for _, m := range lst {
		login := strings.ToLower(m.Login)
		org.Add(login)
	}
}

func (o *OrgLazyLoad) GetPeople() (set.Set, error) {
	var errs error
	if o.Init {
		return o.People, o.Err
	}
	org := set.Empty()
	for login := range o.Input {
		lst, err := memberExpansion(o.Ctx, o.User, o.Caps, o.Repo, login)
		if err != nil {
			msg := fmt.Sprintf("Attempting to expand %s", login)
			errs = multierror.Append(errs, exterror.Append(err, msg))
		} else {
			if lst != nil {
				addToOrg(lst, org)
			} else {
				org.Add(strings.ToLower(login))
			}
		}
	}
	o.People = org
	o.Err = errs
	o.Init = true
	return o.People, o.Err
}
