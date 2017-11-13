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
	"testing"

	"github.com/capitalone/checks-out/cache"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/set"

	"github.com/gin-gonic/gin"
)

func parseMaintainerInternal(m *model.Maintainer, t *testing.T) {
	if len(m.RawPeople) != len(people) {
		t.Fatalf("Wanted %d maintainers, got %d", len(people), len(m.RawPeople))
	}
	for _, want := range people {
		got, ok := m.RawPeople[want.Login]
		if !ok {
			t.Errorf("Wanted user %s in file", want.Login)
		} else if want.Login != got.Login {
			t.Errorf("Wanted login %s, got %s", want.Login, got.Login)
		}
	}
}

func TestParseMaintainerText(t *testing.T) {
	var files = []string{maintainerFile, maintainerFileEmail, maintainerFileSimple, maintainerFileMixed}
	for _, file := range files {
		parsed, err := ParseMaintainer(nil, nil, []byte(file), nil, "text")
		if err != nil {
			t.Fatal(err)
		}
		parseMaintainerInternal(parsed, t)
	}
}

func TestParseMaintainerHJSON(t *testing.T) {
	var files = []string{maintainerFileHJSON}
	for _, file := range files {
		parsed, err := ParseMaintainer(nil, nil, []byte(file), nil, "hjson")
		if err != nil {
			t.Fatal(err)
		}
		parseMaintainerInternal(parsed, t)
	}
}

var people = []model.Person{
	{Login: "bradrydzewski"},
	{Login: "mattnorris"},
}

var maintainerFile = `
Brad Rydzewski <brad.rydzewski@mail.com> (@bradrydzewski)
Matt Norris <matt.norris@mail.com> (@mattnorris)
`

var maintainerFileEmail = `
bradrydzewski <brad.rydzewski@mail.com>
mattnorris <matt.norris@mail.com>
`

// simple format with usernames only. includes
// spaces and comments.
var maintainerFileSimple = `
bradrydzewski
mattnorris`

// simple format with usernames only. includes
// spaces and comments.
var maintainerFileMixed = `
bradrydzewski
Matt Norris <matt.norris@mail.com> (@mattnorris)
`

// advanced hjson format for the maintainers file.
var maintainerFileHJSON = `
{
  people:
  {
    bradrydzewski:
    {
      name: Brad Rydzewski
      email: brad.rydzewski@mail.com
      login: bradrydzewski
    }
    mattnorris:
    {
      name: Matt Norris
      email: matt.norris@mail.com
      login: mattnorris
    }
  }
  org:
  {
    core:
    {
      people:
      [
        mattnorris
        bradrydzewski
      ]
    }
  }
}
`

var teamOrgFile = `
github-org illuminati
github-team a-team org1
github-team b-team
`

var teamRepoSelf = `
github-team repo-self
`

var tomlFile = `
[people]
[people.person1]
[people.person2]
login = "person2"
[people.person3]
login = "person3"
[people.person4]
login = "person4"
[people.person5]
login = "person5"
[people.person6]
login = "person6"

[org]
[org.org1]
people = ["person2","person4","person6"]

[org.org2]
people = ["person1","person3","person5"]
`

func TestParseToml(t *testing.T) {
	parsed, err := parseMaintainerToml([]byte(tomlFile))
	if err != nil {
		t.Fatal(err)
	}
	validateToml(parsed, t)
}

func validateToml(parsed *model.Maintainer, t *testing.T) {
	if len(parsed.RawPeople) != 6 {
		t.Errorf("Expected 6 people, got %d", len(parsed.RawPeople))
	}
	p1, ok := parsed.RawPeople["person1"]
	if !ok {
		t.Fatal("Missing person1")
	}
	if p1.Login != "person1" {
		t.Errorf("Person1 login field is incorrect: %s", p1.Login)
	}
	if len(parsed.RawOrg) != 2 {
		t.Errorf("Expected 2 orgs, got %d", len(parsed.RawOrg))
	}
	o1, ok := parsed.RawOrg["org1"]
	if !ok {
		t.Fatal("Missing org1")
	}
	o1Peeps := set.New("person2", "person4", "person6")
	if len(o1.People.Intersection(o1Peeps)) != 3 {
		t.Errorf("Expected %v in org1, had %v", o1Peeps, o1.People)
	}

	o2, ok := parsed.RawOrg["org2"]
	if !ok {
		t.Fatal("Missing org2")
	}
	o2Peeps := set.New("person1", "person3", "person5")
	if len(o2.People.Intersection(o2Peeps)) != 3 {
		t.Errorf("Expected %v in org2, had %v", o2Peeps, o2.People)
	}
}

func TestParseLegacy(t *testing.T) {
	m, err := ParseMaintainer(nil, nil, []byte(tomlFile), nil, "legacy")
	if err != nil {
		t.Fatal(err)
	}
	validateToml(m, t)
	m2, err := ParseMaintainer(nil, nil, []byte(maintainerFile), nil, "legacy")
	if err != nil {
		t.Fatal(err)
	}
	parseMaintainerInternal(m2, t)
}

func TestParseTeamAndOrg(t *testing.T) {
	parsed, err := parseMaintainerText(nil, nil, []byte(teamOrgFile), nil)
	if err != nil {
		t.Fatal(err)
	}
	org := parsed.RawOrg["illuminati"]
	if org == nil {
		t.Error("The illuminati is missing")
	}
	if !org.People.Contains("github-org illuminati") {
		t.Error("The illuminati is missing its members")
	}
	team := parsed.RawOrg["org1-a-team"]
	if team == nil {
		t.Error("The a-team is missing")
	}
	if !team.People.Contains("github-team a-team org1") {
		t.Error("The a-team is missing its members")
	}
	team = parsed.RawOrg["b-team"]
	if team == nil {
		t.Error("The b-team is missing")
	}
	if !team.People.Contains("github-team b-team") {
		t.Error("The b-team is missing its members")
	}
}

func TestParseTeamRepoSelf(t *testing.T) {
	c := &gin.Context{}
	csh := cache.Default()
	csh.Set("teams:foo", set.New())
	cache.ToContext(c, csh)
	repo := &model.Repo{
		Owner: "foo",
		Name:  "bar",
		Org:   true,
	}
	parsed, err := parseMaintainerText(c, nil, []byte(teamRepoSelf), repo)
	if err != nil {
		t.Fatal(err)
	}
	org := parsed.RawOrg["_"]
	if org == nil {
		t.Error("The repo-self org is missing")
	}
	if !org.People.Contains("github-org repo-self") {
		t.Error("The repo-self org is missing its members")
	}
}
