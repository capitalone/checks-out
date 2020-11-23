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
	"github.com/capitalone/checks-out/strings/miniglob"
	"github.com/capitalone/checks-out/strings/rxserde"

	log "github.com/sirupsen/logrus"
)

func fileMatch(globs []miniglob.MiniGlob, filename string) bool {
	for _, g := range globs {
		if g.Regex.MatchString(filename) {
			return true
		}
	}
	return false
}

func matchesRegexp(exprs []rxserde.RegexSerde, candidate string) bool {
	for _, expr := range exprs {
		if expr.Regex.MatchString(candidate) {
			return true
		}
	}
	return false
}

// O(n^2) algorithm. Consider rewriting if performance becomes an issue
func matchesPaths(globs []miniglob.MiniGlob, files []CommitFile) bool {
	if len(files) == 0 {
		return false
	}
	for _, f := range files {
		if !fileMatch(globs, f.Filename) {
			return false
		}
	}
	return true
}

// O(n^2) algorithm. Consider rewriting if performance becomes an issue
func matchesPathsRegexp(exprs []rxserde.RegexSerde, files []CommitFile) bool {
	if len(files) == 0 {
		return false
	}
	for _, f := range files {
		if !matchesRegexp(exprs, f.Filename) {
			return false
		}
	}
	return true
}

func matchesScope(branch *Branch, scope *ApprovalScope, files []CommitFile) bool {
	paths := true
	branches := true
	pathRegexp := true
	baseRegexp := true
	compareRegexp := true
	if len(scope.Paths) > 0 {
		paths = matchesPaths(scope.Paths, files)
	}
	if len(scope.Branches) > 0 {
		branches = scope.Branches.Contains(branch.BaseName)
	}
	if len(scope.PathRegexp) > 0 {
		pathRegexp = matchesPathsRegexp(scope.PathRegexp, files)
	}
	if len(scope.BaseRegexp) > 0 {
		baseRegexp = matchesRegexp(scope.BaseRegexp, branch.BaseName)
	}
	if len(scope.CompareRegexp) > 0 {
		compareRegexp = matchesRegexp(scope.CompareRegexp, branch.CompareName)
	}
	return paths && branches && pathRegexp && baseRegexp && compareRegexp
}

var internalErrorPolicy = ApprovalPolicy{
	Scope: &ApprovalScope{},
	Match: MatcherHolder{&FalseMatch{}},
}

func FindApprovalPolicy(req *ApprovalRequest) *ApprovalPolicy {
	for _, approval := range req.Config.Approvals {
		if matchesScope(&req.PullRequest.Branch, approval.Scope, req.Files) {
			return approval
		}
	}
	log.Warnf("Internal error. repo %s does not have a default scope.",
		req.Repository.Name)
	return &internalErrorPolicy
}
