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
	"github.com/capitalone/checks-out/set"
)

func titleAction(req *ApprovalRequest, f Feedback, proc Processor) bool {
	forbid := req.IsTitleMatch()
	if !forbid {
		proc(f, ValidTitle)
	}
	return forbid
}

func authorLimitAction(req *ApprovalRequest, f Feedback, people set.Set, proc Processor) {
	people.Add(f.GetAuthor().String())
	proc(f, ValidAuthor)
}

func authorPolicyAction(req *ApprovalRequest, f Feedback, people set.Set, proc Processor) {
	people.Add(f.GetAuthor().String())
}

func approvalAction(req *ApprovalRequest, f Feedback, people set.Set, proc Processor) {
	author := f.GetAuthor().String()
	if f.IsApproval(req) && !people.Contains(author) {
		people.Add(author)
		proc(f, Approval)
	}
}

func antiMatch(req *ApprovalRequest, f Feedback, people set.Set, proc Processor) {
	author := f.GetAuthor().String()
	match := f.IsApproval(req)
	antiMatch := f.IsDisapproval(req)
	if match && antiMatch {
		// do nothing if comment indicates approval and disapproval
		return
	} else if antiMatch && !people.Contains(author) {
		people.Add(author)
		proc(f, DisapprovalInsert)
	} else if match && people.Contains(author) {
		people.Remove(author)
		proc(f, DisapprovalRemove)
	}
}

func doMatch(candidates set.Set, self bool, min int, req *ApprovalRequest,
	proc Processor, action MatchAction, feedback []Feedback) bool {

	participants := set.Empty()

	for _, f := range feedback {
		author := f.GetAuthor().String()
		// cannot approve your own pull request
		if !self && author == req.PullRequest.Author.String() {
			continue
		}
		// the user must a member of the candidate set
		if _, ok := candidates[author]; !ok {
			continue
		}
		action(req, f, participants, proc)
	}
	return len(participants) >= min
}

func (match *UniverseMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	candidates := set.Empty()
	for _, f := range feedback {
		candidates.Add(f.GetAuthor().String())
	}
	return doMatch(candidates, match.Self, match.Approvals, req, proc, a, feedback)
}

func (match *MaintainerMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	candidates := set.Empty()
	for k := range req.Maintainer.People {
		candidates.Add(k)
	}
	return doMatch(candidates, match.Self, match.Approvals, req, proc, a, feedback)
}

func (match *AnonymousMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	candidates := set.Empty()
	for entity := range match.Entities {
		ent := entity.String()
		if org, ok := req.Maintainer.Org[ent]; ok {
			candidates.AddAll(org.People)
		} else {
			candidates.Add(ent)
		}
	}
	return doMatch(candidates, match.Self, match.Approvals, req, proc, a, feedback)
}

func (match *EntityMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	candidates := set.Empty()
	ent := match.Entity.String()
	if org, ok := req.Maintainer.Org[ent]; ok {
		candidates = org.People
	} else {
		candidates.Add(ent)
	}
	return doMatch(candidates, match.Self, match.Approvals, req, proc, a, feedback)
}

func (match *UsMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	candidates := set.Empty()
	if orgs, ok := req.Maintainer.PersonToOrg[req.PullRequest.Author.String()]; ok {
		for name := range orgs {
			if org, ok := req.Maintainer.Org[name]; ok {
				candidates.AddAll(org.People)
			}
		}
	}
	return doMatch(candidates, match.Self, match.Approvals, req, proc, a, feedback)
}

func (match *ThemMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	us := set.Empty()
	if orgs, ok := req.Maintainer.PersonToOrg[req.PullRequest.Author.String()]; ok {
		for name := range orgs {
			if org, ok := req.Maintainer.Org[name]; ok {
				us.AddAll(org.People)
			}
		}
	}
	all := set.Empty()
	for k := range req.Maintainer.People {
		all.Add(k)
	}
	candidates := all.Difference(us)
	return doMatch(candidates, match.Self, match.Approvals, req, proc, a, feedback)
}

func (match *AtLeastMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	if len(match.Choose) == 0 {
		return false
	}
	count := 0
	for _, m := range match.Choose {
		if m.Match(req, proc, a, feedback) {
			count++
		}
	}
	return count >= match.Approvals
}

func (match *AuthorMatch) Match(req *ApprovalRequest, proc Processor, _ MatchAction, feedback []Feedback) bool {
	authorReq := *req
	authorComment := Comment{Author: req.PullRequest.Author}
	authorReq.ApprovalComments = []Feedback{&authorComment}
	return match.Inner.Match(&authorReq, proc, authorPolicyAction, authorReq.ApprovalComments)
}

func (match *IssueAuthorMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	candidates := set.Empty()
	self := req.PullRequest.Author
	for _, issue := range req.Issues {
		author := issue.Author
		if self != author {
			candidates.Add(author.String())
		}
	}
	return doMatch(candidates, false, len(candidates), req, proc, a, feedback)
}

func (match *AndMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	if len(match.And) == 0 {
		return false
	}
	result := true
	for _, m := range match.And {
		// prevent short circuit evaluation
		result = m.Match(req, proc, a, feedback) && result
	}
	return result
}

func (match *OrMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	if len(match.Or) == 0 {
		return false
	}
	result := false
	for _, m := range match.Or {
		// prevent short circuit evaluation
		result = m.Match(req, proc, a, feedback) || result
	}
	return result
}

func (match *NotMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	return !match.Not.Match(req, proc, a, feedback)
}

func (match *TrueMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	return true
}

func (match *FalseMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	return false
}

func (match *DisableMatch) Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool {
	return true
}

func (match *DisableMatch) ChangePolicy(policy *ApprovalPolicy) {
	if policy.Merge == nil {
		m := DefaultMerge()
		policy.Merge = &m
	}
	if policy.Tag == nil {
		tag := DefaultTag()
		policy.Tag = &tag
	}
	policy.AntiMatch.Matcher = &FalseMatch{}
}

func (match *MaintainerMatch) GetType() string {
	return "all"
}

func (match *UniverseMatch) GetType() string {
	return "universe"
}

func (match *UsMatch) GetType() string {
	return "us"
}

func (match *ThemMatch) GetType() string {
	return "them"
}

func (match *AnonymousMatch) GetType() string {
	return "anonymous"
}

func (match *EntityMatch) GetType() string {
	return "entity"
}

func (match *NotMatch) GetType() string {
	return "not"
}

func (match *AndMatch) GetType() string {
	return "and"
}

func (match *OrMatch) GetType() string {
	return "or"
}

func (match *AtLeastMatch) GetType() string {
	return "atleast"
}

func (match *AuthorMatch) GetType() string {
	return "author"
}

func (match *IssueAuthorMatch) GetType() string {
	return "issue-author"
}

func (match *TrueMatch) GetType() string {
	return "true"
}

func (match *FalseMatch) GetType() string {
	return "false"
}

func (match *DisableMatch) GetType() string {
	return "off"
}
