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
	"strconv"

	"github.com/capitalone/checks-out/matcher"
	"github.com/capitalone/checks-out/set"
	"github.com/capitalone/checks-out/strings/lowercase"

	"github.com/mspiegel/go-multierror"
	"github.com/pkg/errors"
)

func GenerateMatcher(input string) (Matcher, error) {
	tokens := matcher.BuildTokens(input)
	parseTree, err := matcher.BuildParseTree(tokens)
	if err != nil {
		return nil, err
	}

	return walkTree(parseTree)
}

func buildAnonymousMatcher(pt *matcher.AnonymousParseToken) (Matcher, error) {
	m := DefaultAnonymousMatch()
	m.Entities = set.NewLowerFromString(pt.Members...)
	err := parseCommonMatch(&m.CommonMatch, pt.Attributes)
	return m, err
}

func buildNounMatcher(pt *matcher.NounParseToken) (Matcher, error) {
	// figure out if this is a:
	// org name
	// special org name (us, them, anyone, people)
	// boolean (true, false)
	var m Matcher
	var err error
	switch pt.Name {
	case "off":
		if len(pt.Attributes) > 0 {
			return nil, errors.New("Attributes are not allowed for off")
		}
		m = &DisableMatch{}
	case "true":
		if len(pt.Attributes) > 0 {
			return nil, errors.New("Attributes are not allowed for true")
		}
		m = &TrueMatch{}
	case "false":
		if len(pt.Attributes) > 0 {
			return nil, errors.New("Attributes are not allowed for false")
		}
		m = &FalseMatch{}
	case "issue-author":
		m = &IssueAuthorMatch{}
	case "all":
		d := DefaultMaintainerMatch()
		m = d
		err = parseCommonMatch(&d.CommonMatch, pt.Attributes)
	case "universe":
		d := DefaultUniverseMatch()
		m = d
		err = parseCommonMatch(&d.CommonMatch, pt.Attributes)
	case "us":
		d := DefaultUsMatch()
		m = d
		err = parseCommonMatch(&d.CommonMatch, pt.Attributes)
	case "them":
		d := DefaultThemMatch()
		m = d
		err = parseCommonMatch(&d.CommonMatch, pt.Attributes)
	default:
		d := DefaultEntityMatch()
		d.Entity = lowercase.Create(pt.Name)
		m = d
		err = parseCommonMatch(&d.CommonMatch, pt.Attributes)
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func buildAndMatcher(pt *matcher.AndOrParseToken) (Matcher, error) {
	cm := &AndMatch{}
	child, err := walkTree(pt.Left)
	if err != nil {
		return nil, err
	}
	cm.And = append(cm.And, MatcherHolder{Matcher: child})
	child, err = walkTree(pt.Right)
	if err != nil {
		return nil, err
	}
	cm.And = append(cm.And, MatcherHolder{Matcher: child})
	return cm, nil
}

func buildOrMatcher(pt *matcher.AndOrParseToken) (Matcher, error) {
	cm := &OrMatch{}
	child, err := walkTree(pt.Left)
	if err != nil {
		return nil, err
	}
	cm.Or = append(cm.Or, MatcherHolder{Matcher: child})
	child, err = walkTree(pt.Right)
	if err != nil {
		return nil, err
	}
	cm.Or = append(cm.Or, MatcherHolder{Matcher: child})
	return cm, nil
}

func buildAndOrMatcher(pt *matcher.AndOrParseToken) (Matcher, error) {
	if pt.JKind == matcher.JOINER_AND {
		return buildAndMatcher(pt)
	} else if pt.JKind == matcher.JOINER_OR {
		return buildOrMatcher(pt)
	} else {
		return nil, errors.Errorf("Unknown operator %v", pt)
	}
}

func buildNotMatcher(pt *matcher.NotParseToken) (Matcher, error) {
	m := &NotMatch{}
	child, err := walkTree(pt.Child)
	if err != nil {
		return nil, err
	}
	m.Not = MatcherHolder{Matcher: child}
	return m, nil
}

func buildFuncMatcher(pt *matcher.FunctionParseToken) (Matcher, error) {
	switch pt.Name {
	case "atleast":
		return buildAtLeastMatcher(pt)
	case "author":
		return buildAuthorMatcher(pt)
	default:
		return nil, errors.Errorf("Unknown function '%s'", pt.Name)
	}
}

func getSymbol(pt matcher.ParseToken) (string, error) {
	switch noun := pt.(type) {
	case *matcher.NounParseToken:
		return noun.Name, nil
	default:
		return "", errors.Errorf("Unable to convert %v to a symbol", pt)
	}
}

func buildAtLeastMatcher(pt *matcher.FunctionParseToken) (Matcher, error) {
	var errs error
	m := &AtLeastMatch{}
	if len(pt.Parameters) == 0 {
		return nil, errors.New("atleast() function must have at least one argument")
	}
	sym, err := getSymbol(pt.Parameters[0])
	if err != nil {
		return nil, err
	}
	count, valid := strconv.Atoi(sym)
	if valid != nil {
		return nil, errors.Errorf("atleast() function first argument expected number, observed %s", pt.Parameters[0])
	}
	m.Approvals = count
	for _, e := range pt.Parameters[1:] {
		c, err := walkTree(e)
		if err != nil {
			errs = multierror.Append(errs, err)
		} else {
			m.Choose = append(m.Choose, MatcherHolder{Matcher: c})
		}
	}
	return m, errs
}

func buildAuthorMatcher(pt *matcher.FunctionParseToken) (Matcher, error) {
	m := &AuthorMatch{}
	if len(pt.Parameters) != 1 {
		return nil, errors.New("author() function must have one argument")
	}
	inner, err := walkTree(pt.Parameters[0])
	if err != nil {
		return nil, err
	}
	m.Inner = MatcherHolder{Matcher: inner}
	return m, nil
}

func parseCommonMatch(m *CommonMatch, attributes map[string]string) error {
	maxAllowed := 0
	if approvals, ok := attributes["count"]; ok {
		count, valid := strconv.Atoi(approvals)
		if valid != nil {
			return errors.Errorf("Expected number, found %s for count attribute on %+v", approvals, m)
		}
		m.Approvals = count
		maxAllowed++
	}
	if self, ok := attributes["self"]; ok {
		s, valid := strconv.ParseBool(self)
		if valid != nil {
			return errors.Errorf("Expected true or false, found %s for self attribute on %+v", self, m)
		}
		m.Self = s
		maxAllowed++
	}
	if len(attributes) > maxAllowed {
		return errors.Errorf("Unexpected attributes found on %+v", m)
	}
	return nil
}

func walkTree(root matcher.ParseToken) (Matcher, error) {
	if root == nil {
		return &TrueMatch{}, nil
	}
	switch pt := root.(type) {
	case *matcher.NounParseToken:
		return buildNounMatcher(pt)
	case *matcher.AnonymousParseToken:
		return buildAnonymousMatcher(pt)
	case *matcher.AndOrParseToken:
		return buildAndOrMatcher(pt)
	case *matcher.NotParseToken:
		return buildNotMatcher(pt)
	case *matcher.FunctionParseToken:
		return buildFuncMatcher(pt)
	default:
		return nil, errors.Errorf("Unknown parse token type: %+v", root)
	}
}
