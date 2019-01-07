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
	"regexp"
	"testing"

	"github.com/capitalone/checks-out/strings/miniglob"
	"github.com/capitalone/checks-out/strings/rxserde"

	"github.com/capitalone/checks-out/set"
)

func TestMatchesBranch(t *testing.T) {
	req := createRequest()
	emptyScope := ApprovalScope{}
	if !matchesScope(&req.PullRequest.Branch, &emptyScope, []CommitFile{}) {
		t.Error("Empty scope should match against request")
	}
	masterScope := ApprovalScope{Branches: set.New("master")}
	if !matchesScope(&req.PullRequest.Branch, &masterScope, []CommitFile{}) {
		t.Error("Master scope should match against request")
	}
	foobarScope := ApprovalScope{Branches: set.New("foobar")}
	if matchesScope(&req.PullRequest.Branch, &foobarScope, []CommitFile{}) {
		t.Error("Foobar scope should not match against request")
	}
}

func TestMatchesPaths(t *testing.T) {
	var files []CommitFile
	files = append(files, CommitFile{
		Filename: "a",
	})
	files = append(files, CommitFile{
		Filename: "b",
	})
	files = append(files, CommitFile{
		Filename: "foo/bar",
	})
	globs := []miniglob.MiniGlob{
		miniglob.MustCreate("a"),
		miniglob.MustCreate("b"),
		miniglob.MustCreate("foo/*"),
	}
	if !matchesPaths(globs, files) {
		t.Error("Path policy did not match")
	}
	globs = []miniglob.MiniGlob{
		miniglob.MustCreate("a"),
		miniglob.MustCreate("b"),
		miniglob.MustCreate("/**r/"),
	}
	if !matchesPaths(globs, files) {
		t.Error("Path policy did not match")
	}
	globs = []miniglob.MiniGlob{
		miniglob.MustCreate("a"),
		miniglob.MustCreate("foo/*"),
	}
	if matchesPaths(globs, files) {
		t.Error("Path policy should not match")
	}
	globs = []miniglob.MiniGlob{
		miniglob.MustCreate("a"),
		miniglob.MustCreate("b"),
		miniglob.MustCreate("foo/bar/baz"),
	}
	if matchesPaths(globs, files) {
		t.Error("Path policy should not match")
	}
}

func TestMatchesPathsRegexp(t *testing.T) {
	var files []CommitFile
	files = append(files, CommitFile{
		Filename: "a",
	})
	files = append(files, CommitFile{
		Filename: "b",
	})
	files = append(files, CommitFile{
		Filename: "foo/bar",
	})
	globs := []rxserde.RegexSerde{
		rxserde.RegexSerde{Regex: regexp.MustCompile("^a$")},
		rxserde.RegexSerde{Regex: regexp.MustCompile("^b$")},
		rxserde.RegexSerde{Regex: regexp.MustCompile("foo/.*")},
	}
	if !matchesPathsRegexp(globs, files) {
		t.Error("Path policy did not match")
	}
	globs = []rxserde.RegexSerde{
		rxserde.RegexSerde{Regex: regexp.MustCompile("^a$")},
		rxserde.RegexSerde{Regex: regexp.MustCompile("^b$")},
		rxserde.RegexSerde{Regex: regexp.MustCompile("bar")},
	}
	if !matchesPathsRegexp(globs, files) {
		t.Error("Path policy did not match")
	}
	globs = []rxserde.RegexSerde{
		rxserde.RegexSerde{Regex: regexp.MustCompile("^a$")},
		rxserde.RegexSerde{Regex: regexp.MustCompile("foo/.*")},
	}
	if matchesPathsRegexp(globs, files) {
		t.Error("Path policy should not match")
	}
	globs = []rxserde.RegexSerde{
		rxserde.RegexSerde{Regex: regexp.MustCompile("^a$")},
		rxserde.RegexSerde{Regex: regexp.MustCompile("^b$")},
		rxserde.RegexSerde{Regex: regexp.MustCompile("foo/bar/baz")},
	}
	if matchesPathsRegexp(globs, files) {
		t.Error("Path policy should not match")
	}
}

//req.Config.Approvals
//req.PullRequest.Branch.BaseName
//req.PullRequest.Branch.CompareName
//req.Files
//req.Repository.Name
func TestFindApprovalPolicy(t *testing.T) {
	req := &ApprovalRequest{
		Config: &Config{
			Approvals: []*ApprovalPolicy{
				//todo
			},
		},
		PullRequest: &PullRequest{
			Branch: Branch{
				BaseName:    "", //todo
				CompareName: "", //todo
			},
		},
		Files: []CommitFile{
			//todo
		},
		Repository: &Repo{
			Name: "", //todo
		},
	}
	policy := FindApprovalPolicy(req)
	if policy == nil {
		t.Error("didn't find the monorepo policy")
	}
}
