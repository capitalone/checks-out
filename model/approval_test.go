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
package model

import (
	"encoding/json"
	"fmt"
	"github.com/capitalone/checks-out/strings/lowercase"
	"github.com/capitalone/checks-out/strings/rxserde"
	"reflect"
	"regexp"
	"testing"

	"github.com/capitalone/checks-out/hjson"
	"github.com/capitalone/checks-out/set"
)

var peopleApproval = `"{Foo,bar,baz,foo}"`

var atLeastApproval = `"atleast(10, a, b, c or d)"`

var nestedApproval = `
approvals: [
{
  match: "{Foo,bar,baz,foo} and a"
}
]
`

func TestAnonymousMatchUnmarshal(t *testing.T) {
	var container MatcherHolder
	err := hjson.Unmarshal([]byte(peopleApproval), &container)
	if err != nil {
		t.Fatal("Error parsing anonymous approval match", err)
	}
	o1 := container.Matcher.(*AnonymousMatch).Entities
	e1 := set.NewLowerFromString("foo", "bar", "baz")
	if !reflect.DeepEqual(o1, e1) {
		t.Error("Anonymous set generated incorrectly", o1)
	}
}

func TestAtLeastMatcherUnmarshal(t *testing.T) {
	var container MatcherHolder
	err := hjson.Unmarshal([]byte(atLeastApproval), &container)
	if err != nil {
		t.Fatal("Error parsing atleast approval match", err)
	}
	s := container.Matcher.(*AtLeastMatch)
	if len(s.Choose) != 3 {
		t.Error("Atleast list generated incorrectly", s.Choose)
	}
	if s.Approvals != 10 {
		t.Error("Approval count parsed incorrectly", s.Approvals)
	}
}

var implicitMatcher = `
approvals: [
  {
	  scope: {
		  branches: [ "master" ]
	  }
	}
	{
	  scope: {
		  branches: []
		}
  }
]
`

func TestImplicitApproval(t *testing.T) {
	c, err := ParseConfig([]byte(implicitMatcher), AllowAll())
	if err != nil {
		t.Fatalf("Error parsing configuration file: %+v", err)
	}
	t0 := reflect.TypeOf(c.Approvals[0].Match.Matcher).String()
	t1 := reflect.TypeOf(c.Approvals[1].Match.Matcher).String()
	if t0 != "*model.MaintainerMatch" {
		t.Errorf("First approval match is not maintainers: %s", t0)
	}
	if t1 != "*model.MaintainerMatch" {
		t.Errorf("Second approval match is not maintainers: %s", t1)
	}
}

func TestMarshal(t *testing.T) {
	m := MatcherHolder{Matcher: DefaultMatcher()}
	text, err := json.Marshal(m)
	if err != nil {
		t.Fatal("Error marshaling match policy", err)
	}
	var mjson string
	err = json.Unmarshal(text, &mjson)
	if err != nil {
		t.Fatal("Error unmarshaling match policy", err)
	}
	if mjson != "all[count=1,self=true]" {
		t.Errorf("Unexpected marshaling, got %s, expected all[count=1,self=true]", mjson)
	}
}

func TestOffMatch(t *testing.T) {
	a := `
	{
		"match": "off"
	}
	`
	var ap ApprovalPolicy

	err := json.Unmarshal([]byte(a), &ap)
	if err != nil {
		t.Fatal("Error unmarshalling approval policy", err)
	}

	setupPolicyDefaults(0, &ap)

	if _, ok := ap.Match.Matcher.(*DisableMatch); !ok {
		t.Fatalf("Expected DisableMatch, got %+v", ap.Match.Matcher)
	}
	if ap.AntiMatch.Matcher == nil {
		t.Fatal("Expected non-nil anti-matcher")
	}
	if _, ok := ap.AntiMatch.Matcher.(*FalseMatch); !ok {
		t.Fatalf("Expected FalseMatch for anti-matcher, got %+v", ap.AntiMatch.Matcher)
	}
	if ap.Merge == nil {
		t.Fatal("Expected non-nil merge rule")
	}
	if ap.Tag == nil {
		t.Fatal("Expected non-nil tag rule")
	}
}

func TestTrueMatchNils(t *testing.T) {
	a := `
	{
		"match": "true"
	}
	`
	var ap ApprovalPolicy

	err := json.Unmarshal([]byte(a), &ap)
	if err != nil {
		t.Fatal("Error unmarshalling approval policy", err)
	}

	setupPolicyDefaults(0, &ap)

	if _, ok := ap.Match.Matcher.(*TrueMatch); !ok {
		t.Fatalf("Expected TrueMatch, got %+v", ap.Match.Matcher)
	}
	if ap.Merge != nil {
		t.Fatal("Expected nil merge rule")
	}
	if ap.Tag != nil {
		t.Fatal("Expected nil tag rule")
	}
}

func TestRoundTripMatcher(t *testing.T) {
	m := `
	{
		"match": "atleast(2, foo[count=3,self=true],all[count=2,self=false] or not universe and baz[count=5], true or false and fred[self=true])"
	}
	`
	var ap ApprovalPolicy

	err := json.Unmarshal([]byte(m), &ap)
	if err != nil {
		t.Fatal("Error unmarshalling approval policy", err)
	}
	roundTrip, err := json.Marshal(ap)
	if err != nil {
		t.Fatal("Error marshaling approval policy", err)
	}
	expected := `{"match":"atleast(2,foo[count=3,self=true],all[count=2,self=false] or not universe[count=1,self=true] and baz[count=5,self=true],true or false and fred[count=1,self=true])"}`
	if string(roundTrip) != expected {
		t.Fatalf("Error round-tripping approval policy. Expected '%s', got '%s'", expected, string(roundTrip))
	}
}

func TestAnonymousGroupRoundTripping(t *testing.T) {
	vals := map[string]string{
		`{"match": "{bob,jane}"}`:                                `{"match":"{bob,jane}[count=1,self=true]"}`,
		`{"match": "{bob,jane}[count=2,self=false]"}`:            `{"match":"{bob,jane}[count=2,self=false]"}`,
		`{"match": "atleast(1,{bob,jane}[count=2,self=false])"}`: `{"match":"atleast(1,{bob,jane}[count=2,self=false])"}`,
		`{"match": "atleast(1,{bob,jane})"}`:                     `{"match":"atleast(1,{bob,jane}[count=1,self=true])"}`,
	}
	for k, v := range vals {
		var ap ApprovalPolicy

		err := json.Unmarshal([]byte(k), &ap)
		if err != nil {
			t.Fatal("Error unmarshalling approval policy", err)
		}
		roundTrip, err := json.Marshal(ap)
		if err != nil {
			t.Fatal("Error marshaling approval policy", err)
		}
		if string(roundTrip) != v {
			t.Errorf("Error round-tripping approval policy. Expected '%s', got '%s'", v, string(roundTrip))
		}
	}

}

func TestNestedApprovals(t *testing.T) {
	c := DefaultConfig()
	c.Approvals = []*ApprovalPolicy{
		{
			Name:     "nested",
			Position: 1,
			Match: MatcherHolder{
				Matcher: &TrueMatch{},
			},
			Scope: &ApprovalScope{
				Branches: set.Set{
					"dev": true,
				},
				Nested: []InnerScope{
					{
						PathRegexp: rxserde.RegexSerde{
							Regex: regexp.MustCompile(".*/gui/.*"),
						},
						Match: MatcherHolder{
							Matcher: &EntityMatch{
								Entity: lowercase.Create("front_end_devs"),
								CommonMatch: CommonMatch{
									Approvals: 2,
									Self:      false,
								},
							},
						},
					},
					{
						PathRegexp: rxserde.RegexSerde{
							Regex: regexp.MustCompile(".*/server/.*"),
						},
						Match: MatcherHolder{
							Matcher: &EntityMatch{
								Entity: lowercase.Create("back_end_devs"),
								CommonMatch: CommonMatch{
									Approvals: 2,
									Self:      false,
								},
							},
						},
					},
				},
			},
		},
	}
	js, err := hjson.Marshal(c)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(js))
}
