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
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/capitalone/checks-out/set"
	"github.com/capitalone/checks-out/strings/lowercase"
	"github.com/capitalone/checks-out/strings/miniglob"
	"github.com/capitalone/checks-out/strings/rxserde"

	"github.com/pkg/errors"
)

type ApprovalRequest struct {
	Config              *Config
	Maintainer          *MaintainerSnapshot
	PullRequest         *PullRequest
	Repository          *Repo
	Issues              []*Issue
	ApprovalComments    []Feedback
	DisapprovalComments []Feedback
	Files               []CommitFile
}

type Feedback interface {
	IsApproval(req *ApprovalRequest) bool
	IsDisapproval(req *ApprovalRequest) bool
	GetAuthor() lowercase.String
	GetBody() string
	GetSubmittedAt() time.Time
}

// Processor updates bookkepping for comment tracking.
type Processor func(Feedback, ApprovalOp)

// MatchAction decides whether to invoke the Processor
type MatchAction func(*ApprovalRequest, Feedback, set.Set, Processor)

func Approve(request *ApprovalRequest, policy *ApprovalPolicy, p Processor) bool {
	authorRequest := *request
	authorComment := Comment{Author: request.PullRequest.Author}
	authorRequest.ApprovalComments = []Feedback{&authorComment}

	if titleAction(&authorRequest, &authorComment, p) {
		return false
	}

	if !policy.AuthorMatch.Match(&authorRequest, p, authorLimitAction, authorRequest.ApprovalComments) {
		return false
	}

	if policy.AntiMatch.Match(request, p, antiMatch, request.DisapprovalComments) {
		return false
	}

	success := policy.Match.Match(request, p, approvalAction, request.ApprovalComments)
	return success
}

// ApprovalPolicy combines an approval scope (that determines
// when the policy can be applied) and an approval match
// (that determines whether the match is successful)
type ApprovalPolicy struct {
	// Name is used to describe this policy to humans.
	// Optional but if specified it must be unique.
	Name string `json:"name,omitempty"`
	// Position is the 1-based index of this policy in the approval array.
	Position int `json:"-"`
	// ApprovalScope determines when the policy can be applied
	Scope *ApprovalScope `json:"scope,omitempty"`
	// Match is a JSON object that stores the approval match.
	// every approval policy needs a Match
	Match MatcherHolder `json:"match"`
	// AntiMatch is a JSON object that stores the disapproval match.
	AntiMatch *MatcherHolder `json:"antimatch,omitempty"`
	// AntiMatch is a JSON object that stores the author match.
	AuthorMatch *MatcherHolder `json:"authormatch,omitempty"`
	// Tag is an optional tag section to be used for this
	// approval scope. If this field is empty then the
	// global tag section is used.
	Tag *TagConfig `json:"tag,omitempty"`
	// Merge is an optional merge section to be used for this
	// approval scope. If this field is empty then the
	// global merge section is used.
	Merge *MergeConfig `json:"merge,omitempty"`
	// Pattern is an optional regular expression to be used for this
	// approval scope. If this field is empty then the
	// global pattern is used.
	Pattern *rxserde.RegexSerde `json:"pattern,omitempty"`
	// AntiPattern is an optional regular expression to be used for this
	// approval scope. If this field is empty then the
	// global antipattern is used.
	AntiPattern *rxserde.RegexSerde `json:"antipattern,omitempty"`
	// AntiTitle is an optional regular expression to be used for this
	// approval scope. If this field is empty then the
	// global antititle is used.
	AntiTitle *rxserde.RegexSerde `json:"antititle,omitempty"`
	// Feedback is an optional feedback section to be used for this
	// approval scope. If this field is empty then the
	// global feedback section is used.
	Feedback *FeedbackConfig `json:"feedback,omitempty"`
}

// ApprovalScope determines when the policy can be applied
type ApprovalScope struct {
	Paths         []miniglob.MiniGlob  `json:"paths,omitempty"`
	Branches      set.Set              `json:"branches,omitempty"`
	PathRegexp    []rxserde.RegexSerde `json:"regexpaths,omitempty"`
	BaseRegexp    []rxserde.RegexSerde `json:"regexbase,omitempty"`
	CompareRegexp []rxserde.RegexSerde `json:"regexcompare,omitempty"`
}

// MatcherHolder stores an an Matcher
// JSON marshal and unmarshal are implemented
// on this struct.
type MatcherHolder struct {
	Matcher
}

// Matcher determines whether the match is successful)
type Matcher interface {
	Match(req *ApprovalRequest, proc Processor, a MatchAction, feedback []Feedback) bool
	GetType() string
	Validate(m *MaintainerSnapshot) error
}

type ChangePolicy interface {
	ChangePolicy(policy *ApprovalPolicy)
}

type CommonMatch struct {
	// minimum number of approvals required
	Approvals int `json:"approvals"`
	// if true then author can self-approve request
	Self bool `json:"self"`
}

// UniverseMatch accepts the request when the number
// of people who have approved is greater than or
// equal to the threshold. Approvals are not restricted
// to project maintainers.
type UniverseMatch struct {
	CommonMatch
}

func (match UniverseMatch) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`"universe[count=%d,self=%v]"`, match.Approvals, match.Self)
	return []byte(s), nil
}

// MaintainerMatch accepts the request when the number
// of people who have approved is greater than or
// equal to the threshold. Approvals are restricted
// to project maintainers.
type MaintainerMatch struct {
	CommonMatch
}

func (match MaintainerMatch) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`"all[count=%d,self=%v]"`, match.Approvals, match.Self)
	return []byte(s), nil
}

// AnonymousMatch accepts the request when the number
// of people who have approved is greater than or
// equal to the threshold.
type AnonymousMatch struct {
	CommonMatch
	// restrict the approvers to this set
	Entities set.LowerSet `json:"entities"`
}

func (match AnonymousMatch) MarshalJSON() ([]byte, error) {
	p := match.Entities.Keys().ToStringSlice()
	sort.Strings(p) //make sure they are always in the same order so we can test
	s := fmt.Sprintf(`"{%s}[count=%d,self=%v]"`, strings.Join(p, ","), match.Approvals, match.Self)
	return []byte(s), nil
}

// EntityMatch accepts the request when the
// group or person meets the minimum number of approvals.
type EntityMatch struct {
	CommonMatch
	// name of the group or person to match
	Entity lowercase.String `json:"entity"`
}

func (match EntityMatch) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`"%s[count=%d,self=%v]"`, match.Entity, match.Approvals, match.Self)
	return []byte(s), nil
}

// UsMatch is restricted to maintainers who share
// a group with the author of the pull request.
type UsMatch struct {
	CommonMatch
}

func (match UsMatch) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`"us[count=%d,self=%v]"`, match.Approvals, match.Self)
	return []byte(s), nil
}

// ThemMatch is restricted to maintainers who
// do not share a group with the author of the
// pull request.
type ThemMatch struct {
	CommonMatch
}

func (match ThemMatch) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`"them[count=%d,self=%v]"`, match.Approvals, match.Self)
	return []byte(s), nil
}

type AtLeastMatch struct {
	// minimum number of approvals required
	Approvals int             `json:"approvals"`
	Choose    []MatcherHolder `json:"choose"`
}

func (match AtLeastMatch) MarshalJSON() ([]byte, error) {
	c := make([]string, 0, len(match.Choose))
	for _, v := range match.Choose {
		b, e := json.Marshal(v)
		if e != nil {
			return nil, e
		}
		c = append(c, string(b[1:len(b)-1]))
	}
	s := fmt.Sprintf(`"atleast(%d,%s)"`, match.Approvals, strings.Join(c, ","))
	return []byte(s), nil
}

type AuthorMatch struct {
	Inner MatcherHolder `json:"inner"`
}

func (match AuthorMatch) MarshalJSON() ([]byte, error) {
	b, e := json.Marshal(match.Inner)
	if e != nil {
		return nil, e
	}
	c := string(b[1 : len(b)-1])
	s := fmt.Sprintf(`"author(%s)"`, c)
	return []byte(s), nil
}

type IssueAuthorMatch struct{}

func (match IssueAuthorMatch) MarshalJSON() ([]byte, error) {
	return []byte("issue-author"), nil
}

// AndMatch performs a boolean 'and' operation on
// two or more Matchers.
type AndMatch struct {
	And []MatcherHolder `json:"and"`
}

func (match AndMatch) MarshalJSON() ([]byte, error) {
	c := make([]string, 0, len(match.And))
	for _, v := range match.And {
		b, e := json.Marshal(v)
		if e != nil {
			return nil, e
		}
		c = append(c, string(b[1:len(b)-1]))
	}
	s := strings.Join(c, " and ")
	return []byte(fmt.Sprintf(`"%s"`, s)), nil
}

// OrMatch performs a boolean 'or' operation on
// two or more Matchers.
type OrMatch struct {
	Or []MatcherHolder `json:"or"`
}

func (match OrMatch) MarshalJSON() ([]byte, error) {
	c := make([]string, 0, len(match.Or))
	for _, v := range match.Or {
		b, e := json.Marshal(v)
		if e != nil {
			return nil, e
		}
		c = append(c, string(b[1:len(b)-1]))
	}
	s := strings.Join(c, " or ")
	return []byte(fmt.Sprintf(`"%s"`, s)), nil
}

// NotMatch performs a boolean 'not' operation on
// a matcher.
type NotMatch struct {
	Not MatcherHolder `json:"not"`
}

func (match NotMatch) MarshalJSON() ([]byte, error) {
	b, e := json.Marshal(match.Not)
	if e != nil {
		return nil, e
	}
	return []byte(fmt.Sprintf(`"not %s"`, string(b[1:len(b)-1]))), nil
}

// TrueMatch always returns true.
type TrueMatch struct{}

func (match TrueMatch) MarshalJSON() ([]byte, error) {
	return []byte(`"true"`), nil
}

// FalseMatch always returns false.
type FalseMatch struct{}

func (match FalseMatch) MarshalJSON() ([]byte, error) {
	return []byte(`"false"`), nil
}

// DisableMatch always returns true and disables service.
type DisableMatch struct{}

func (match DisableMatch) MarshalJSON() ([]byte, error) {
	return []byte(`"off"`), nil
}

func (match MatcherHolder) MarshalJSON() ([]byte, error) {
	return json.Marshal(match.Matcher)
}

func (match *MatcherHolder) UnmarshalJSON(data []byte) error {
	var text string
	err := json.Unmarshal(data, &text)
	if err != nil {
		return err
	}
	m, err := GenerateMatcher(text)
	if err != nil {
		return errors.Errorf("Unable to parse approval policy '%s'. %s",
			text, err.Error())
	}
	match.Matcher = m
	return nil
}

func DefaultApprovalPolicy() *ApprovalPolicy {
	a := new(ApprovalPolicy)
	a.AntiMatch = new(MatcherHolder)
	a.AuthorMatch = new(MatcherHolder)
	a.Scope = DefaultApprovalScope()
	a.Match.Matcher = DefaultMatcher()
	a.AntiMatch.Matcher = DefaultMatcher()
	a.AuthorMatch.Matcher = DefaultUniverseMatch()
	return a
}

func DefaultApprovalScope() *ApprovalScope {
	return &ApprovalScope{}
}

func DefaultMatcher() Matcher {
	return DefaultMaintainerMatch()
}

func DefaultMaintainerMatch() *MaintainerMatch {
	m := new(MaintainerMatch)
	m.Approvals = 1
	m.Self = true
	return m
}

func DefaultUniverseMatch() *UniverseMatch {
	m := new(UniverseMatch)
	m.Approvals = 1
	m.Self = true
	return m
}

func DefaultAnonymousMatch() *AnonymousMatch {
	m := new(AnonymousMatch)
	m.Approvals = 1
	m.Self = true
	return m
}

func DefaultEntityMatch() *EntityMatch {
	m := new(EntityMatch)
	m.Approvals = 1
	m.Self = true
	return m
}

func DefaultUsMatch() *UsMatch {
	m := new(UsMatch)
	m.Approvals = 1
	m.Self = true
	return m
}

func DefaultThemMatch() *ThemMatch {
	m := new(ThemMatch)
	m.Approvals = 1
	m.Self = true
	return m
}

// IsTitleMatch returns true if the text matches the
// antititle regular expression.
func (req *ApprovalRequest) IsTitleMatch() bool {
	var regExp *regexp.Regexp
	policy := FindApprovalPolicy(req)
	if policy.AntiTitle != nil {
		regExp = policy.AntiTitle.Regex
	} else if req.Config.AntiTitle != nil {
		regExp = req.Config.AntiTitle.Regex
	}
	if regExp == nil {
		// title matching is optional
		return false
	}
	return regExp.MatchString(req.PullRequest.Title)
}
