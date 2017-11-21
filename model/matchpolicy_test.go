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
	"reflect"
	"regexp"
	"testing"

	"github.com/capitalone/checks-out/set"
	"github.com/capitalone/checks-out/strings/lowercase"
	"github.com/capitalone/checks-out/strings/rxserde"
)

func createRequest() *ApprovalRequest {
	request := &ApprovalRequest{}
	request.Maintainer = &MaintainerSnapshot{}
	request.PullRequest = &PullRequest{}
	request.Maintainer.People = map[string]*Person{
		"alice": &Person{Name: "alice"},
		"bob":   &Person{Name: "bob"},
		"carol": &Person{Name: "carol"},
		"dan":   &Person{Name: "dan"},
	}
	request.Maintainer.Org = map[string]Org{
		"guelph":     &OrgSerde{People: map[string]bool{"alice": true, "bob": true}},
		"ghibelline": &OrgSerde{People: map[string]bool{"carol": true, "dan": true}},
	}
	request.PullRequest.Author = lowercase.Create("alice")
	request.PullRequest.Branch.BaseName = "master"
	request.ApprovalComments = []Feedback{
		&Comment{Author: lowercase.Create("Alice"), Body: "I approve"},
		&Comment{Author: lowercase.Create("Bob"), Body: "I approve"},
		&Comment{Author: lowercase.Create("bob"), Body: "I approve"},
		&Comment{Author: lowercase.Create("CAROL"), Body: "I approve"},
		&Comment{Author: lowercase.Create("mystery"), Body: "mystery"},
		&Review{Author: lowercase.Create("dan"), State: lowercase.Create("approved")},
		&Comment{Author: lowercase.Create("foobar"), Body: "I approve"},
	}
	request.DisapprovalComments = request.ApprovalComments
	request.Config = NonEmptyConfig()
	request.Issues = []*Issue{
		&Issue{Author: lowercase.Create("Alice")},
		&Issue{Author: lowercase.Create("Bob")},
	}
	return request
}

func TestAnyoneMatch(t *testing.T) {
	request := createRequest()
	universe := UniverseMatch{}
	universe.Approvals = 5
	universe.Self = true
	match := MatcherHolder{&universe}
	request.Config.Approvals[0].Match = match
	approvers := set.Empty()
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == false {
		t.Error("match failure")
	}
	if !reflect.DeepEqual(approvers, set.New("alice", "bob", "carol", "dan", "foobar")) {
		t.Error("Approvers set generated incorrectly", approvers)
	}
}

func TestAuthorPolicy(t *testing.T) {
	request := createRequest()
	m := AnonymousMatch{}
	m.Approvals = 1
	m.Self = true
	m.Entities = set.NewLowerFromString("alice", "bob", "foo", "bar")
	author := AuthorMatch{}
	author.Inner = MatcherHolder{Matcher: &m}
	request.Config.Approvals[0].Match = MatcherHolder{Matcher: &author}
	approvers := set.Empty()
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == false {
		t.Error("match failure")
	}
	if len(approvers) != 0 {
		t.Error("approvers set should be empty")
	}
	m = AnonymousMatch{}
	m.Approvals = 1
	m.Self = true
	m.Entities = set.NewLowerFromString("bob", "foo", "bar")
	author = AuthorMatch{}
	author.Inner = MatcherHolder{Matcher: &m}
	request.Config.Approvals[0].Match = MatcherHolder{Matcher: &author}
	approvers = set.Empty()
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == true {
		t.Error("match success")
	}
	if len(approvers) != 0 {
		t.Error("approvers set should be empty")
	}
}

func TestAnonymousMatch(t *testing.T) {
	request := createRequest()
	m := AnonymousMatch{}
	m.Approvals = 4
	m.Self = true
	m.Entities = set.NewLowerFromString("foo", "bar", "baz")
	match := MatcherHolder{&m}
	request.Config.Approvals[0].Match = match
	approvers := set.Empty()
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == true {
		t.Error("match did not fail")
	}
	m.Entities = set.NewLowerFromString("alice", "bob", "carol", "dan", "foo", "bar")
	request.Config.Approvals[0].Match = match
	approvers = set.Empty()
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == false {
		t.Error("match did not succeed")
	}
	if !reflect.DeepEqual(approvers, set.New("alice", "bob", "carol", "dan")) {
		t.Error("Approvers set generated incorrectly", approvers)
	}
}

func TestAnonymousUserAndGroupMatch(t *testing.T) {
	request := createRequest()
	m := AnonymousMatch{}
	m.Approvals = 4
	m.Self = true
	m.Entities = set.NewLowerFromString("guelph", "alice")
	match := MatcherHolder{&m}
	request.Config.Approvals[0].Match = match
	approvers := set.Empty()
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == true {
		t.Error("match did not fail")
	}
	m.Entities = set.NewLowerFromString("guelph", "alice", "carol", "dan", "foo", "bar")
	request.Config.Approvals[0].Match = match
	approvers = set.Empty()
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == false {
		t.Error("match did not succeed")
	}
	if !reflect.DeepEqual(approvers, set.New("alice", "bob", "carol", "dan")) {
		t.Error("Approvers set generated incorrectly", approvers)
	}
}

func TestIssueAuthorMatch(t *testing.T) {
	request := createRequest()
	m := IssueAuthorMatch{}
	request.Config.Approvals[0].Match = MatcherHolder{&m}
	approvers := set.Empty()
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == false {
		t.Error("match did not succeed")
	}
	request.Issues = append(request.Issues, &Issue{Author: lowercase.Create("david")})
	approvers = set.Empty()
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == true {
		t.Error("match did not fail")
	}
}

func TestAuthorMatch(t *testing.T) {
	request := createRequest()
	m := MaintainerMatch{}
	a := AnonymousMatch{}
	a.Approvals = 1
	a.Entities = set.NewLowerFromString("bob")
	a.Self = true
	m.Approvals = 4
	m.Self = true
	request.Config.Approvals[0].Match = MatcherHolder{&m}
	request.Config.Approvals[0].AuthorMatch = &MatcherHolder{&a}
	approvers := set.Empty()
	author := false
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		} else if op == ValidAuthor {
			author = true
		}
	})
	if success == true {
		t.Error("match did not fail")
	}
	if author == true {
		t.Error("author match did not fail")
	}
	a.Entities = set.NewLowerFromString("alice", "bob")
	approvers = set.Empty()
	author = false
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		} else if op == ValidAuthor {
			author = true
		}
	})
	if success == false {
		t.Error("match did not succeed")
	}
	if author == false {
		t.Error("author match did not succeed")
	}
}

func TestMaintainersMatch(t *testing.T) {
	request := createRequest()
	m := MaintainerMatch{}
	m.Approvals = 5
	m.Self = true
	match := MatcherHolder{&m}
	request.Config.Approvals[0].Match = match
	approvers := set.Empty()
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == true {
		t.Error("match did not fail")
	}
	m.Approvals = 4
	m.Self = true
	match = MatcherHolder{&m}
	request.Config.Approvals[0].Match = match
	approvers = set.Empty()
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == false {
		t.Error("match did not succeed")
	}
	if !reflect.DeepEqual(approvers, set.New("alice", "bob", "carol", "dan")) {
		t.Error("Approvers set generated incorrectly", approvers)
	}
}

func TestEntityGroupMatch(t *testing.T) {
	request := createRequest()
	m := EntityMatch{}
	m.Entity = lowercase.Create("guelph")
	m.Approvals = 2
	m.Self = false
	match := MatcherHolder{&m}
	request.Config.Approvals[0].Match = match
	approvers := set.Empty()
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == true {
		t.Error("match did not fail")
	}
	m.Approvals = 2
	m.Self = true
	match = MatcherHolder{&m}
	request.Config.Approvals[0].Match = match
	approvers = set.Empty()
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == false {
		t.Error("match did not succeed")
	}
	if !reflect.DeepEqual(approvers, set.New("alice", "bob")) {
		t.Error("Approvers set generated incorrectly", approvers)
	}
}

func TestEntityIndividualMatch(t *testing.T) {
	request := createRequest()
	m := EntityMatch{}
	m.Entity = lowercase.Create("alice")
	m.Approvals = 1
	m.Self = false
	match := MatcherHolder{&m}
	request.Config.Approvals[0].Match = match
	approvers := set.Empty()
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == true {
		t.Error("match did not fail")
	}
	m.Approvals = 1
	m.Self = true
	match = MatcherHolder{&m}
	request.Config.Approvals[0].Match = match
	approvers = set.Empty()
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == false {
		t.Error("match did not succeed")
	}
	if !reflect.DeepEqual(approvers, set.New("alice")) {
		t.Error("Approvers set generated incorrectly", approvers)
	}
}

func TestUsMatch(t *testing.T) {
	request := createRequest()
	m := UsMatch{}
	m.Approvals = 3
	m.Self = true
	match := MatcherHolder{&m}
	request.Config.Approvals[0].Match = match
	approvers := set.Empty()
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == true {
		t.Error("match did not fail")
	}
	m.Approvals = 2
	m.Self = true
	match = MatcherHolder{&m}
	request.Config.Approvals[0].Match = match
	approvers = set.Empty()
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == false {
		t.Error("match did not succeed")
	}
	if !reflect.DeepEqual(approvers, set.New("alice", "bob")) {
		t.Error("Approvers set generated incorrectly", approvers)
	}
}

func TestThemMatch(t *testing.T) {
	request := createRequest()
	m := ThemMatch{}
	m.Approvals = 3
	m.Self = true
	match := MatcherHolder{&m}
	request.Config.Approvals[0].Match = match
	approvers := set.Empty()
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == true {
		t.Error("match did not fail")
	}
	m.Approvals = 2
	m.Self = true
	match = MatcherHolder{&m}
	request.Config.Approvals[0].Match = match
	approvers = set.Empty()
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		}
	})
	if success == false {
		t.Error("match did not succeed")
	}
	if !reflect.DeepEqual(approvers, set.New("carol", "dan")) {
		t.Error("Approvers set generated incorrectly", approvers)
	}
}

func TestTrueMatch(t *testing.T) {
	request := createRequest()
	match := MatcherHolder{&TrueMatch{}}
	request.Config.Approvals[0].Match = match
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, _ ApprovalOp) {})
	if success == false {
		t.Error("match did not succeed")
	}
}

func TestDisableMatch(t *testing.T) {
	request := createRequest()
	match := MatcherHolder{&DisableMatch{}}
	request.Config.Approvals[0].Match = match
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, _ ApprovalOp) {})
	if success == false {
		t.Error("match did not succeed")
	}
}

func TestFalseMatch(t *testing.T) {
	request := createRequest()
	match := MatcherHolder{&FalseMatch{}}
	request.Config.Approvals[0].Match = match
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, _ ApprovalOp) {})
	if success == true {
		t.Error("match did not fail")
	}
}

func TestNotMatch(t *testing.T) {
	request := createRequest()
	match := MatcherHolder{&NotMatch{Not: MatcherHolder{&FalseMatch{}}}}
	request.Config.Approvals[0].Match = match
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, _ ApprovalOp) {})
	if success == false {
		t.Error("match did not succeed")
	}
}

func TestAndMatch(t *testing.T) {
	request := createRequest()
	match := MatcherHolder{
		&AndMatch{And: []MatcherHolder{
			{&FalseMatch{}},
			{&TrueMatch{}}}}}
	request.Config.Approvals[0].Match = match
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, _ ApprovalOp) {})
	if success == true {
		t.Error("match did not fail")
	}
	match = MatcherHolder{
		&AndMatch{And: []MatcherHolder{
			{&TrueMatch{}},
			{&TrueMatch{}}}}}
	request.Config.Approvals[0].Match = match
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, _ ApprovalOp) {})
	if success == false {
		t.Error("match did not succeed")
	}
}

func TestOrMatch(t *testing.T) {
	request := createRequest()
	match := MatcherHolder{
		&OrMatch{Or: []MatcherHolder{
			{&FalseMatch{}},
			{&FalseMatch{}}}}}
	request.Config.Approvals[0].Match = match
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, _ ApprovalOp) {})
	if success == true {
		t.Error("match did not fail")
	}
	match = MatcherHolder{
		&OrMatch{Or: []MatcherHolder{
			MatcherHolder{&FalseMatch{}},
			MatcherHolder{&TrueMatch{}}}}}
	request.Config.Approvals[0].Match = match
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, _ ApprovalOp) {})
	if success == false {
		t.Error("match did not succeed")
	}
}

func TestDisapprovalComments(t *testing.T) {
	request := createRequest()
	request.ApprovalComments[3] = &Comment{Author: lowercase.Create("carol"), Body: "NONONO"}
	request.Config.AntiPattern = &rxserde.RegexSerde{Regex: regexp.MustCompile("NONONO")}
	universe := MaintainerMatch{}
	universe.Approvals = 4
	universe.Self = true
	match := MatcherHolder{&universe}
	request.Config.Approvals[0].Match = match
	request.Config.Approvals[0].AntiMatch = &match
	approvers := set.Empty()
	disapprovers := set.Empty()

	//copy of closure from web/approval.go
	approverFunc := func(f Feedback, op ApprovalOp) {
		author := f.GetAuthor().String()
		switch op {
		case Approval:
			approvers.Add(author)
		case DisapprovalInsert:
			disapprovers.Add(author)
		case DisapprovalRemove:
			disapprovers.Remove(author)
		case ValidAuthor:
			// do nothing
		case ValidTitle:
			// do nothing
		default:
			t.Fatalf("Unknown approval operation %d", op)
		}
	}
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, approverFunc)
	if success == true {
		t.Error("match did not fail")
	}
	if !reflect.DeepEqual(approvers, set.New("alice", "bob", "dan")) {
		t.Error("Approvers set generated incorrectly", approvers)
	}
	if !reflect.DeepEqual(disapprovers, set.New("carol")) {
		t.Error("disapprovers set generated incorrectly", disapprovers)
	}
	request.ApprovalComments = append(request.ApprovalComments, &Comment{Author: lowercase.Create("carol"), Body: "I approve"})
	request.DisapprovalComments = request.ApprovalComments
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, approverFunc)
	if success == false {
		t.Error("match did not succeed")
	}
	if !reflect.DeepEqual(approvers, set.New("alice", "bob", "carol", "dan")) {
		t.Error("Approvers set generated incorrectly", approvers)
	}
	if !reflect.DeepEqual(disapprovers, set.New()) {
		t.Error("disapprovers set generated incorrectly", disapprovers)
	}
}

func TestDisapprovalReview(t *testing.T) {
	request := createRequest()
	request.ApprovalComments[3] = &Review{Author: lowercase.Create("carol"), State: lowercase.Create("changes_requested")}
	universe := MaintainerMatch{}
	universe.Approvals = 4
	universe.Self = true
	match := MatcherHolder{&universe}
	request.Config.Approvals[0].Match = match
	request.Config.Approvals[0].AntiMatch = &match
	approvers := set.Empty()
	disapprovers := set.Empty()

	//copy of closure from web/approval.go
	approverFunc := func(f Feedback, op ApprovalOp) {
		author := f.GetAuthor().String()
		switch op {
		case Approval:
			approvers.Add(author)
		case DisapprovalInsert:
			disapprovers.Add(author)
		case DisapprovalRemove:
			disapprovers.Remove(author)
		case ValidAuthor:
			// do nothing
		case ValidTitle:
			// do nothing
		default:
			t.Fatalf("Unknown approval operation %d", op)
		}
	}
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, approverFunc)
	if success == true {
		t.Error("match did not fail")
	}
	if !reflect.DeepEqual(approvers, set.New("alice", "bob", "dan")) {
		t.Error("Approvers set generated incorrectly", approvers)
	}
	if !reflect.DeepEqual(disapprovers, set.New("carol")) {
		t.Error("disapprovers set generated incorrectly", disapprovers)
	}
	request.ApprovalComments = append(request.ApprovalComments, &Comment{Author: lowercase.Create("carol"), Body: "I approve"})
	request.DisapprovalComments = request.ApprovalComments
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, approverFunc)
	if success == false {
		t.Error("match did not succeed")
	}
	if !reflect.DeepEqual(approvers, set.New("alice", "bob", "carol", "dan")) {
		t.Error("Approvers set generated incorrectly", approvers)
	}
	if !reflect.DeepEqual(disapprovers, set.New()) {
		t.Error("disapprovers set generated incorrectly", disapprovers)
	}
}

func TestTitleMatch(t *testing.T) {
	request := createRequest()
	request.PullRequest = &PullRequest{}
	request.PullRequest.Title = "WIP: An experiment"
	request.Config.AntiTitle = &rxserde.RegexSerde{Regex: regexp.MustCompile("^WIP")}
	m := AnonymousMatch{}
	m.Approvals = 4
	m.Self = true
	m.Entities = set.NewLowerFromString("foo", "bar", "baz")
	match := MatcherHolder{&m}
	m.Entities = set.NewLowerFromString("alice", "bob", "carol", "dan", "foo", "bar")
	request.Config.Approvals[0].Match = match
	approvers := set.Empty()
	validTitle := false
	policy := FindApprovalPolicy(request)
	success, _ := Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		} else if op == ValidTitle {
			validTitle = true
		}
	})
	if success == true {
		t.Error("match did not fail")
	}
	if validTitle == true {
		t.Error("validTitle did not fail")
	}
	request.PullRequest.Title = "An experiment"
	validTitle = false
	policy = FindApprovalPolicy(request)
	success, _ = Approve(request, policy, func(f Feedback, op ApprovalOp) {
		if op == Approval {
			approvers.Add(f.GetAuthor().String())
		} else if op == ValidTitle {
			validTitle = true
		}
	})
	if success == false {
		t.Error("match did not succeed")
	}
	if validTitle == false {
		t.Error("validTitle did not succeed")
	}
}
