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
	"testing"

	"github.com/capitalone/checks-out/cache"
	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/set"
	"github.com/capitalone/checks-out/store"
	"github.com/capitalone/checks-out/strings/lowercase"

	"github.com/gin-gonic/gin"
)

func init() {
	cache.Longterm.Set("people:Foo", &model.Person{Login: "Foo"})
	cache.Longterm.Set("people:Bar", &model.Person{Login: "Bar"})
	cache.Longterm.Set("people:Baz", &model.Person{Login: "Baz"})
}

type mockRemote struct {
	remote.Remote
}

func (m *mockRemote) GetOrgMembers(c context.Context, user *model.User, org string) (set.Set, error) {
	switch org {
	case "a":
		return map[string]bool{
			"Foo": true,
		}, nil
	case "org":
		return map[string]bool{
			"Foo": true,
			"Bar": true,
		}, nil
	default:
		return map[string]bool{}, nil
	}
}

func (m *mockRemote) GetTeamMembers(c context.Context, user *model.User, org string, team string) (set.Set, error) {
	return map[string]bool{
		"Bar": true,
		"Baz": true,
	}, nil
}

func TestMaintainerToSnapshot(t *testing.T) {
	c := &gin.Context{}
	u := &model.User{}
	caps := model.AllowAll()
	r := &model.Repo{Org: true}
	m := &model.Maintainer{
		RawPeople: map[string]*model.Person{
			"Foo": &model.Person{Login: "Foo", Name: "Mr. Foo"},
		},
		RawOrg: map[string]*model.OrgSerde{
			"G1": &model.OrgSerde{People: set.New("github-org a")},
			"G2": &model.OrgSerde{People: set.New("github-org org")},
			"G3": &model.OrgSerde{People: set.New("github-team team")},
		},
	}
	remote.ToContext(c, &mockRemote{})
	s, err := maintainerToSnapshot(c, u, caps, r, m)
	if err != nil {
		t.Error("Error converting maintainer to snapshot", err)
	}
	if len(s.People) != 3 {
		t.Error("people not populated")
	}
	if _, ok := s.Org["G1"]; ok {
		t.Error("group G1 was not converted to lowercase")
	}
	g1, err := s.Org["g1"].GetPeople()
	if err != nil {
		t.Error(err)
	}
	if g1.Contains("Foo") {
		t.Error("user Foo was not converted to lowercase")
	}
	if !g1.Contains("foo") {
		t.Error("user foo is missing from group g1")
	}
	if s.People["foo"].Name != "Mr. Foo" {
		t.Error("foo user was overridden", s.People["foo"])
	}
}

func TestMaintainerToSnapshotOrgSelf(t *testing.T) {
	c := &gin.Context{}
	u := &model.User{}
	caps := model.AllowAll()
	r := &model.Repo{Owner: "a", Org: true}
	m := &model.Maintainer{
		RawOrg: map[string]*model.OrgSerde{
			"g1": &model.OrgSerde{People: set.New("github-org repo-self")},
		},
	}
	remote.ToContext(c, &mockRemote{})
	s, err := maintainerToSnapshot(c, u, caps, r, m)
	if err != nil {
		t.Error("Error converting maintainer to snapshot", err)
	}
	if _, ok := s.People["foo"]; !ok {
		t.Error("people not populated")
	}
	if _, ok := s.People["bar"]; ok {
		t.Error("bar is populated")
	}
	g1, err := s.Org["g1"].GetPeople()
	if err != nil {
		t.Error(err)
	}
	if !g1.Contains("foo") {
		t.Error("org contents not populated")
	}
}

func TestSnapshotValidateApproval(t *testing.T) {
	c := &gin.Context{}
	u := &model.User{}
	caps := model.AllowAll()
	r := &model.Repo{Owner: "a", Org: true}
	m := &model.Maintainer{
		RawOrg: map[string]*model.OrgSerde{
			"g1": &model.OrgSerde{People: set.New("github-org repo-self")},
		},
	}
	remote.ToContext(c, &mockRemote{})
	s, err := maintainerToSnapshot(c, u, caps, r, m)
	if err != nil {
		t.Error("Error converting maintainer to snapshot", err)
	}
	config := model.NonEmptyConfig()
	m1 := model.DefaultEntityMatch()
	m1.Entity = lowercase.Create("foobar")
	config.Approvals[0].Match = model.MatcherHolder{Matcher: m1}
	err = validateSnapshot(config, s)
	if err == nil {
		t.Error("Expected error validating snapshot")
	}
	m2 := model.DefaultAnonymousMatch()
	m2.Entities = set.NewLower(lowercase.Create("foobar"))
	config.Approvals[0].Match = model.MatcherHolder{Matcher: m2}
	err = validateSnapshot(config, s)
	if err == nil {
		t.Error("Expected error validating snapshot")
	}
}

func TestAddMembers(t *testing.T) {
	c := model.OrgSerde{People: set.New("foo")}
	m := model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"foo": &model.Person{Login: "foo", Name: "Mr. Foo"},
		},
		Org: map[string]model.Org{
			"c": &c,
		},
	}
	lst := []*model.Person{
		&model.Person{Login: "foo"},
		&model.Person{Login: "bar"},
		&model.Person{Login: "baz"},
	}
	addToPeople(lst, &m)
	addToOrg(lst, c.People)
	if len(m.People) != 3 {
		t.Error("Persons not added to snapshot", m.People)
	}
	if len(c.People) != 3 {
		t.Error("Persons not added to organization", c.People)
	}
	if m.People["foo"].Name != "Mr. Foo" {
		t.Error("foo user was overridden", m.People["foo"])
	}
}

func TestPopulatePersonToOrg(t *testing.T) {
	m := model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"foo": &model.Person{Login: "foo"},
			"bar": &model.Person{Login: "bar"},
			"baz": &model.Person{Login: "baz"},
		},
		Org: map[string]model.Org{
			"a": &model.OrgSerde{People: set.New("foo")},
			"b": &model.OrgSerde{People: set.New("foo", "bar")},
			"c": &model.OrgSerde{People: set.New("foo", "bar", "baz", "quux")},
		},
	}
	mapping, err := m.PersonToOrg()
	if err != nil {
		t.Error(err)
	}
	if len(mapping) != 3 {
		t.Errorf("foo was not mapped to the correct organizations: %v",
			mapping["foo"])
	}
	if len(mapping["bar"]) != 2 {
		t.Errorf("bar was not mapped to the correct organizations: %v",
			mapping["bar"])
	}
	if len(mapping["baz"]) != 1 {
		t.Errorf("baz was not mapped to the correct organizations: %v",
			mapping["baz"])
	}
	if len(mapping["quux"]) != 0 {
		t.Errorf("quux was not mapped to the correct organizations: %v",
			mapping["quux"])
	}
}

type mockStore struct {
	store.Store
}

func (ms mockStore) GetSlackUrl(hostname string, user string) (string, error) {
	if user == "" && hostname == "floopy-server.slack.com" {
		return "http://this-is-the-url.com", nil
	}
	if user == "jon" && hostname == "bloopy-server.slack.com" {
		return "http://this-is-the-other-url.com", nil
	}
	return "", nil
}

func TestFixSlackTargets(t *testing.T) {
	envvars.Env.Slack.TargetUrl = "http://standard-url.com"
	c := store.AddToContext(context.Background(), mockStore{})
	config := &model.Config{
		Comment: model.CommentConfig{
			Enable: true,
			Targets: []model.TargetConfig{
				{
					Target: "floopy-server.slack.com",
					Names:  []string{"#channel1"},
				},
				{
					Target: "github",
				},
				{
					Target: "slack",
					Names:  []string{"#channelZ"},
				},
				{
					Target: "bloopy-server.slack.com",
					Names:  []string{"#channel2"},
				},
			},
		},
	}
	err := FixSlackTargets(c, config, "jon")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if config.Comment.Targets[0].Target != model.Slack.String() {
		t.Errorf("Unexpected target %s", config.Comment.Targets[0].Target)
	}
	if config.Comment.Targets[0].Url != "http://this-is-the-url.com" {
		t.Errorf("Unexpected url %s", config.Comment.Targets[0].Url)
	}
	if config.Comment.Targets[1].Target != model.Github.String() {
		t.Errorf("Unexpected target %s", config.Comment.Targets[0].Target)
	}
	if config.Comment.Targets[2].Target != model.Slack.String() {
		t.Errorf("Unexpected target %s", config.Comment.Targets[0].Target)
	}
	if config.Comment.Targets[2].Url != "http://standard-url.com" {
		t.Errorf("Unexpected url %s", config.Comment.Targets[0].Url)
	}
	if config.Comment.Targets[3].Target != model.Slack.String() {
		t.Errorf("Unexpected target %s", config.Comment.Targets[0].Target)
	}
	if config.Comment.Targets[3].Url != "http://this-is-the-other-url.com" {
		t.Errorf("Unexpected url %s", config.Comment.Targets[0].Url)
	}
}
