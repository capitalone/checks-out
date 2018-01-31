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
	"bufio"
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
)

func parseMaintainerText(c context.Context, user *model.User, data []byte, r *model.Repo) (*model.Maintainer, error) {
	m := new(model.Maintainer)
	m.RawPeople = map[string]*model.Person{}
	m.RawOrg = map[string]*model.OrgSerde{}

	buf := bytes.NewBuffer(data)
	reader := bufio.NewReader(buf)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		item := parseln(string(line))
		if len(item) == 0 {
			continue
		}

		if name := ParseOrgName(item); len(name) > 0 {
			if name == SelfRepo {
				name = r.Owner
			}
			err := addOrg(m, item, name)
			if err != nil {
				return nil, err
			}
			continue
		}
		if collab := ParseCollabName(item); len(collab) > 0 {
			if collab == SelfRepo {
				collab = r.Slug
			}
			collab += "-collaborators"
			err := addOrg(m, item, collab)
			if err != nil {
				return nil, err
			}
			continue
		}
		if team, org0 := ParseTeamName(item); len(team) > 0 {
			if team == SelfTeam {
				eagerLoad := "github-org " + org0
				org := org0
				if org0 == "" {
					if !r.Org {
						err := fmt.Errorf("Cannot expand GitHub teams for user repository %s/%s",
							r.Owner, r.Name)
						return nil, badRequest(err)
					}
					org = r.Owner
					eagerLoad = "github-org repo-self"
				}
				err := addOrg(m, eagerLoad, "_"+org0)
				if err != nil {
					return nil, err
				}
				teams, err := remote.ListTeams(c, user, org)
				if err != nil {
					return nil, err
				}
				for t := range teams {
					i := fmt.Sprintf("github-team %s", t)
					name := t
					if org0 != "" {
						i = fmt.Sprintf("github-team %s %s", t, org)
						name = org + "-" + t
					}
					err := addOrg(m, i, name)
					if err != nil {
						return nil, err
					}
				}
			} else {
				name := team
				if org0 != "" {
					name = org0 + "-" + team
				}
				err := addOrg(m, item, name)
				if err != nil {
					return nil, err
				}
			}
			continue
		}

		person := parseLogin(item)
		if person == nil {
			person = parseLoginMeta(item)
		}
		if person == nil {
			person = parseLoginEmail(item)
		}
		if person == nil {
			err := fmt.Errorf("Unable to parse line: %s", item)
			return nil, badRequest(err)
		}
		m.RawPeople[person.Login] = person
	}
	return m, nil
}

func addOrg(m *model.Maintainer, item string, name string) error {
	name = strings.ToLower(name)
	o := &model.OrgSerde{People: map[string]bool{item: true}}
	if _, ok := m.RawOrg[name]; ok {
		err := fmt.Errorf("Duplicate organization detected %s", name)
		return badRequest(err)
	}
	m.RawOrg[name] = o
	return nil
}

func parseln(s string) string {
	if s == "" || string(s[0]) == "#" {
		return ""
	}
	index := strings.Index(s, " #")
	if index > -1 {
		s = strings.TrimSpace(s[0:index])
	}
	return s
}

// regular expression determines if a line in the maintainers
// is a reference to an organization.
var reOrg = regexp.MustCompile(`^github-org (\S+)`)

// regular expression determines if a line in the maintainers
// is a reference to collaborators list.
var reCollab = regexp.MustCompile(`^github-collab (\S+)`)

// regular expression determines if a line in the maintainers
// is a reference to a team.
var reTeam = regexp.MustCompile(`^github-team (\S+)\s*(\S*)`)

// regular expression determines if a line in the maintainers
// file only has the single GitHub username and no other metadata.
var reLogin = regexp.MustCompile(`^\S+$`)

// regular expression determines if a line in the maintainers
// file has the username and metadata.
var reLoginMeta = regexp.MustCompile(`(.+) <(.+)> \(@(.+)\)`)

// regular expression determines if a line in the maintainers
// file has the username and email.
var reLoginEmail = regexp.MustCompile(`(.+) <(.+)>`)

func parseLoginMeta(line string) *model.Person {
	matches := reLoginMeta.FindStringSubmatch(line)
	if len(matches) != 4 {
		return nil
	}
	return &model.Person{
		Name:  strings.TrimSpace(matches[1]),
		Email: strings.TrimSpace(matches[2]),
		Login: strings.TrimSpace(matches[3]),
	}
}

func parseLoginEmail(line string) *model.Person {
	matches := reLoginEmail.FindStringSubmatch(line)
	if len(matches) != 3 {
		return nil
	}
	return &model.Person{
		Login: strings.TrimSpace(matches[1]),
		Email: strings.TrimSpace(matches[2]),
	}
}

func parseLogin(line string) *model.Person {
	line = strings.TrimSpace(line)
	if !reLogin.MatchString(line) {
		return nil
	}
	return &model.Person{
		Login: line,
	}
}

func ParseOrgName(line string) string {
	matches := reOrg.FindStringSubmatch(line)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

func ParseCollabName(line string) string {
	matches := reCollab.FindStringSubmatch(line)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

func ParseTeamName(line string) (string, string) {
	matches := reTeam.FindStringSubmatch(line)
	if len(matches) != 3 {
		return "", ""
	}
	return matches[1], matches[2]
}
