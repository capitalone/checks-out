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
package web

import (
	"context"
	"errors"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/notifier"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/set"
	"github.com/capitalone/checks-out/strings/lowercase"
	"github.com/capitalone/checks-out/strings/rxserde"

	"github.com/gin-gonic/gin"
	version "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
)

func TestHandleTimestampMillis(t *testing.T) {
	c := &model.TagConfig{Alg: "timestamp-millis"}
	stamp, err := handleTimestamp(c)
	if err != nil {
		t.Errorf("didn't expect error: %v", err)
	}
	m := time.Now().UTC().Unix()
	m1, err := strconv.ParseInt(stamp, 10, 64)
	if err != nil {
		t.Error(err)
	}
	if math.Abs(float64(m-m1)) > 100 {
		t.Errorf("shouldn't be that different: %d, %d", m, m1)
	}
}

func TestHandleTimestamp3339(t *testing.T) {
	c := &model.TagConfig{Alg: "timestamp-rfc3339"}
	stamp, err := handleTimestamp(c)
	if err != nil {
		t.Errorf("didn't expect error: %v", err)
	}
	//should be able to parse with rfc3339
	t2, err := time.Parse(modifiedRFC3339, stamp)
	if err != nil {
		t.Error(err)
	}
	round := t2.Format(modifiedRFC3339)
	if round != stamp {
		t.Errorf("Expected to be same, but wasn't %s, %s", stamp, round)
	}
}

type myR struct {
	remote.Remote
}

func (m *myR) ListTags(c context.Context, u *model.User, r *model.Repo) ([]model.Tag, error) {
	return []model.Tag{
		"a",
		"0.1.0",
		"0.0.1",
	}, nil
}

func (m *myR) GetCommentsSinceHead(c context.Context, u *model.User, r *model.Repo, num int, noUIMerge bool) ([]*model.Comment, error) {
	return []*model.Comment{
		{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		{
			Author: lowercase.Create("not_test_guy"),
			Body:   "this is not an I approve comment",
		},
		{
			Author: lowercase.Create("not_test_guy"),
			Body:   "I approve",
		},
		{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		{
			Author: lowercase.Create("test_guy2"),
			Body:   "I approve",
		},
		{
			Author: lowercase.Create("test_guy3"),
			Body:   "I approve version:0.0.1",
		},
	}, nil
}

func TestGetMaxVersionComment(t *testing.T) {
	c := &gin.Context{}

	remote.ToContext(c, &myR{})
	config := model.NonEmptyConfig()
	config.Tag.Enable = true
	m := &model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"test_guy": &model.Person{
				Name: "test_guy",
			},
			"test_guy2": &model.Person{
				Name: "test_guy2",
			},
			"test_guy3": &model.Person{
				Name: "test_guy3",
			},
		},
	}
	i := model.PullRequest{
		Issue: model.Issue{Author: lowercase.Create("test_guy")},
	}
	comments := []model.Feedback{
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "this is not an I approve comment",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy2"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy3"),
			Body:   "I approve version:0.0.1",
		},
	}
	request := &model.ApprovalRequest{
		Config:           config,
		Maintainer:       m,
		PullRequest:      &i,
		ApprovalComments: comments,
	}
	ver := getMaxVersionComment(request, model.DefaultApprovalPolicy())
	if ver == nil {
		t.Fatalf("Got nil for version")
	}
	expected, _ := version.NewVersion("0.1.0")
	if !expected.Equal(ver) {
		t.Errorf("Expected %s, got %s", expected.String(), ver.String())
	}
}

func TestGetMaxVersionCommentBadPattern(t *testing.T) {
	c := &gin.Context{}

	remote.ToContext(c, &myR{})
	config := model.NonEmptyConfig()
	config.Tag.Enable = true
	config.Pattern = rxserde.RegexSerde{Regex: nil}
	m := &model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"test_guy": &model.Person{
				Name: "test_guy",
			},
			"test_guy2": &model.Person{
				Name: "test_guy2",
			},
			"test_guy3": &model.Person{
				Name: "test_guy3",
			},
		},
	}
	i := model.PullRequest{
		Issue: model.Issue{Author: lowercase.Create("test_guy")},
	}
	comments := []model.Feedback{
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "this is not an I approve comment",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "not an approval comment",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy2"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy3"),
			Body:   "I approve version:0.0.1",
		},
	}
	request := &model.ApprovalRequest{
		Config:           config,
		Maintainer:       m,
		PullRequest:      &i,
		ApprovalComments: comments,
	}
	ver := getMaxVersionComment(request, model.DefaultApprovalPolicy())
	if ver != nil {
		t.Fatal("Should get nil for version. Version is ", ver)
	}
}

func TestGetMaxVersionCommentNoSelfApproval(t *testing.T) {
	c := &gin.Context{}

	remote.ToContext(c, &myR{})
	policy := model.DefaultApprovalPolicy()
	policy.Match.Matcher.(*model.MaintainerMatch).Self = false
	config := model.NonEmptyConfig()
	config.Approvals[0] = policy
	m := &model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"test_guy": &model.Person{
				Name: "test_guy",
			},
			"test_guy2": &model.Person{
				Name: "test_guy2",
			},
			"test_guy3": &model.Person{
				Name: "test_guy3",
			},
		},
	}
	i := model.PullRequest{
		Issue: model.Issue{Author: lowercase.Create("test_guy")},
	}
	comments := []model.Feedback{
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "this is not an I approve comment",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy2"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy3"),
			Body:   "I approve version:0.0.1",
		},
	}
	request := &model.ApprovalRequest{
		Config:           config,
		Maintainer:       m,
		PullRequest:      &i,
		ApprovalComments: comments,
	}
	ver := getMaxVersionComment(request, policy)
	if ver == nil {
		t.Fatal("Got nil for version")
	}
	expected, _ := version.NewVersion("0.0.1")
	if !expected.Equal(ver) {
		t.Errorf("Expected %s, got %s", expected.String(), ver.String())
	}
}

func TestGetMaxExistingTagFound(t *testing.T) {
	ver := getMaxExistingTag([]model.Tag{
		"a",
		"0.1.0",
		"0.0.1",
	})

	expected, _ := version.NewVersion("0.1.0")
	if !expected.Equal(ver) {
		t.Errorf("Expected %s, got %s", expected.String(), ver.String())
	}
}

func TestGetMaxExistingTagNotFound(t *testing.T) {
	ver := getMaxExistingTag([]model.Tag{
		"a",
		"b",
		"c",
	})

	expected, _ := version.NewVersion("0.0.0")
	if !expected.Equal(ver) {
		t.Errorf("Expected %s, got %s", expected.String(), ver.String())
	}
}

func TestHandleSemver(t *testing.T) {
	c := &gin.Context{}

	remote.ToContext(c, &myR{})
	config := model.NonEmptyConfig()
	config.Tag.Enable = true
	m := &model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"test_guy": &model.Person{
				Name: "test_guy",
			},
			"test_guy2": &model.Person{
				Name: "test_guy2",
			},
			"test_guy3": &model.Person{
				Name: "test_guy3",
			},
		},
	}
	user := &model.User{}
	repo := &model.Repo{}
	hook := &StatusHook{
		Repo: &model.Repo{
			Owner: "test_guy",
			Name:  "test_repo",
		},
	}
	pr := &model.PullRequest{
		Issue: model.Issue{
			Author: lowercase.Create("test_guy"),
		},
	}
	request := &model.ApprovalRequest{
		Config:      config,
		Maintainer:  m,
		PullRequest: pr,
		Repository:  repo,
	}
	ver, err := handleSemver(c, user, hook, request, model.DefaultApprovalPolicy())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	expected, _ := version.NewVersion("0.1.1")
	if expected.String() != ver {
		t.Errorf("Expected %s, got %s", expected.String(), ver)
	}
}

type myR2 struct {
	remote.Remote
}

func (m *myR2) ListTags(c context.Context, u *model.User, r *model.Repo) ([]model.Tag, error) {
	return []model.Tag{
		"a",
		"0.0.1",
		"0.0.2",
	}, nil
}

var testComments = []model.Feedback{
	&model.Comment{
		Author: lowercase.Create("test_guy"),
		Body:   "I approve version:0.1.0",
	},
	&model.Comment{
		Author: lowercase.Create("not_test_guy"),
		Body:   "this is not an I approve comment",
	},
	&model.Comment{
		Author: lowercase.Create("not_test_guy"),
		Body:   "I approve",
	},
	&model.Comment{
		Author: lowercase.Create("test_guy"),
		Body:   "I approve version:0.1.0",
	},
	&model.Comment{
		Author: lowercase.Create("test_guy2"),
		Body:   "I approve",
	},
	&model.Comment{
		Author: lowercase.Create("test_guy3"),
		Body:   "I approve version:0.0.1",
	},
}

func TestHandleSemver2(t *testing.T) {
	c := &gin.Context{}

	remote.ToContext(c, &myR2{})
	config := model.NonEmptyConfig()
	config.Tag.Enable = true
	m := &model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"test_guy": {
				Name: "test_guy",
			},
			"test_guy2": {
				Name: "test_guy2",
			},
			"test_guy3": {
				Name: "test_guy3",
			},
		},
	}
	user := &model.User{}
	repo := &model.Repo{}
	hook := &StatusHook{
		Repo: &model.Repo{
			Owner: "test_guy",
			Name:  "test_repo",
		},
	}
	pr := &model.PullRequest{
		Issue: model.Issue{
			Author: lowercase.Create("test_guy"),
		},
	}
	request := &model.ApprovalRequest{
		Config:              config,
		Maintainer:          m,
		PullRequest:         pr,
		Repository:          repo,
		ApprovalComments:    testComments,
		DisapprovalComments: testComments,
	}
	ver, err := handleSemver(c, user, hook, request, model.DefaultApprovalPolicy())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	expected, _ := version.NewVersion("0.1.0")
	if expected.String() != ver {
		t.Errorf("Expected %s, got %s", expected.String(), ver)
	}
}

type myR3 struct {
	remote.Remote
}

func (m *myR3) ListTags(c context.Context, u *model.User, r *model.Repo) ([]model.Tag, error) {
	return nil, errors.New("This is an error")
}

func (m *myR3) GetCommentsSinceHead(c context.Context, u *model.User, r *model.Repo, num int, noUIMerge bool) ([]*model.Comment, error) {
	return []*model.Comment{
		{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		{
			Author: lowercase.Create("not_test_guy"),
			Body:   "this is not an I approve comment",
		},
		{
			Author: lowercase.Create("not_test_guy"),
			Body:   "I approve",
		},
		{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		{
			Author: lowercase.Create("test_guy2"),
			Body:   "I approve",
		},
		{
			Author: lowercase.Create("test_guy3"),
			Body:   "I approve version:0.0.1",
		},
	}, nil
}

func TestHandleSemver3(t *testing.T) {
	c := &gin.Context{}

	remote.ToContext(c, &myR3{})
	config := model.NonEmptyConfig()
	config.Tag.Enable = true
	m := &model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"test_guy": {
				Name: "test_guy",
			},
			"test_guy2": {
				Name: "test_guy2",
			},
			"test_guy3": {
				Name: "test_guy3",
			},
		},
	}
	user := &model.User{}
	repo := &model.Repo{}
	hook := &StatusHook{
		Repo: &model.Repo{
			Owner: "test_guy",
			Name:  "test_repo",
		},
	}
	pr := &model.PullRequest{
		Issue: model.Issue{
			Author: lowercase.Create("test_guy"),
		},
	}
	request := &model.ApprovalRequest{
		Config:              config,
		Maintainer:          m,
		PullRequest:         pr,
		Repository:          repo,
		ApprovalComments:    testComments,
		DisapprovalComments: testComments,
	}
	ver, err := handleSemver(c, user, hook, request, model.DefaultApprovalPolicy())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	expected, _ := version.NewVersion("0.1.0")
	if expected.String() != ver {
		t.Errorf("Expected %s, got %s", expected.String(), ver)
	}
}

type myRR struct {
	gin.ResponseWriter
}

func (mr *myRR) Header() http.Header {
	return http.Header{}
}

func (my *myRR) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestGetMaxVersionCommentOldPattern(t *testing.T) {
	c := &gin.Context{}

	remote.ToContext(c, &myR{})
	config := model.NonEmptyConfig()

	regex, err := regexp.Compile(`(?i)^I approve\s*(?P<version>\S*)`)
	assert.Nil(t, err)

	config.Pattern = rxserde.RegexSerde{Regex: regex}
	config.Tag.Enable = true
	m := &model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"test_guy": &model.Person{
				Name: "test_guy",
			},
			"test_guy2": &model.Person{
				Name: "test_guy2",
			},
			"test_guy3": &model.Person{
				Name: "test_guy3",
			},
		},
	}
	i := model.PullRequest{
		Issue: model.Issue{Author: lowercase.Create("test_guy")},
	}
	comments := []model.Feedback{
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve 0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "this is not an I approve comment",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve 0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy2"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy3"),
			Body:   "I approve 0.0.1",
		},
	}
	request := &model.ApprovalRequest{
		Config:           config,
		Maintainer:       m,
		PullRequest:      &i,
		ApprovalComments: comments,
	}
	ver := getMaxVersionComment(request, model.DefaultApprovalPolicy())
	if ver == nil {
		t.Fatalf("Got nil for version")
	}
	expected, _ := version.NewVersion("0.1.0")
	if !expected.Equal(ver) {
		t.Errorf("Expected %s, got %s", expected.String(), ver.String())
	}
}

func TestGetCommitComment(t *testing.T) {
	c := &gin.Context{}

	remote.ToContext(c, &myR{})
	config := model.NonEmptyConfig()
	config.Tag.Enable = true
	m := &model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"test_guy": &model.Person{
				Name: "test_guy",
			},
			"test_guy2": &model.Person{
				Name: "test_guy2",
			},
			"test_guy3": &model.Person{
				Name: "test_guy3",
			},
		},
	}
	i := model.PullRequest{
		Issue: model.Issue{Author: lowercase.Create("test_guy")},
	}
	comments := []model.Feedback{
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve comment:Hello",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "this is not an I approve comment",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy2"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy3"),
			Body:   "I approve version:0.0.1 comment:Hi there",
		},
	}
	request := &model.ApprovalRequest{
		Config:              config,
		Maintainer:          m,
		PullRequest:         &i,
		ApprovalComments:    comments,
		DisapprovalComments: comments,
	}
	comment := getCommitComment(request, model.DefaultApprovalPolicy())
	if comment != "Hi there" {
		t.Error("Expected Hi there, got", comment)
	}
}

func TestEligibleForDeletion(t *testing.T) {
	config := model.NonEmptyConfig()
	pr := &model.PullRequest{
		Branch: model.Branch{
			CompareOwner: "foo",
			CompareName:  "baz",
		},
	}
	request := &model.ApprovalRequest{
		Config:      config,
		PullRequest: pr,
		Repository: &model.Repo{
			Owner: "bar",
		},
	}
	mw := &notifier.MessageWrapper{}
	if eligibleForDeletion(request, mw) {
		t.Error("branch belongs to another owner")
	}
	request.Repository.Owner = "foo"
	if eligibleForDeletion(request, mw) {
		t.Error("default policy should not be eligible for deletion")
	}
	config.Approvals[0].Match.Matcher = &model.DisableMatch{}
	if !eligibleForDeletion(request, mw) {
		t.Error("off policy should be eligible for deletion")
	}
	config.Approvals = append(config.Approvals, model.DefaultApprovalPolicy())
	config.Approvals[0].Scope.Branches = set.New("baz")
	if eligibleForDeletion(request, mw) {
		t.Error("branch match should not be eligible for deletion")
	}
}

func TestGetLastVersionComment(t *testing.T) {
	c := &gin.Context{}

	remote.ToContext(c, &myR{})
	config := model.NonEmptyConfig()
	config.Tag.Enable = true
	m := &model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"test_guy": &model.Person{
				Name: "test_guy",
			},
			"test_guy2": &model.Person{
				Name: "test_guy2",
			},
			"test_guy3": &model.Person{
				Name: "test_guy3",
			},
		},
	}
	i := model.PullRequest{
		Issue: model.Issue{Author: lowercase.Create("test_guy")},
	}
	comments := []model.Feedback{
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "this is not an I approve comment",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy2"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy3"),
			Body:   "I approve version:0.0.1",
		},
	}
	request := &model.ApprovalRequest{
		Config:           config,
		Maintainer:       m,
		PullRequest:      &i,
		ApprovalComments: comments,
	}
	ver := getLastVersionComment(request, model.DefaultApprovalPolicy())
	if ver == "" {
		t.Fatalf("Got empty for version")
	}
	if ver != "0.0.1" {
		t.Errorf("Expected 0.0.1, got %s", ver)
	}
}

func TestGetLastVersionCommentBadPattern(t *testing.T) {
	c := &gin.Context{}

	remote.ToContext(c, &myR{})
	config := model.NonEmptyConfig()
	config.Tag.Enable = true
	config.Pattern = rxserde.RegexSerde{Regex: nil}
	m := &model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"test_guy": &model.Person{
				Name: "test_guy",
			},
			"test_guy2": &model.Person{
				Name: "test_guy2",
			},
			"test_guy3": &model.Person{
				Name: "test_guy3",
			},
		},
	}
	i := model.PullRequest{
		Issue: model.Issue{Author: lowercase.Create("test_guy")},
	}
	comments := []model.Feedback{
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "this is not an I approve comment",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "not an approval comment",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy2"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy3"),
			Body:   "I approve version:0.0.1",
		},
	}
	request := &model.ApprovalRequest{
		Config:           config,
		Maintainer:       m,
		PullRequest:      &i,
		ApprovalComments: comments,
	}
	ver := getLastVersionComment(request, model.DefaultApprovalPolicy())
	if ver != "" {
		t.Fatal("Should get empty for version. Version is ", ver)
	}
}

func TestGetLastVersionCommentNoSelfApproval(t *testing.T) {
	c := &gin.Context{}

	remote.ToContext(c, &myR{})
	policy := model.DefaultApprovalPolicy()
	policy.Match.Matcher.(*model.MaintainerMatch).Self = false
	config := model.NonEmptyConfig()
	config.Approvals[0] = policy
	m := &model.MaintainerSnapshot{
		People: map[string]*model.Person{
			"test_guy": &model.Person{
				Name: "test_guy",
			},
			"test_guy2": &model.Person{
				Name: "test_guy2",
			},
			"test_guy3": &model.Person{
				Name: "test_guy3",
			},
		},
	}
	i := model.PullRequest{
		Issue: model.Issue{Author: lowercase.Create("test_guy")},
	}
	comments := []model.Feedback{
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "this is not an I approve comment",
		},
		&model.Comment{
			Author: lowercase.Create("not_test_guy"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy"),
			Body:   "I approve version:0.1.0",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy2"),
			Body:   "I approve",
		},
		&model.Comment{
			Author: lowercase.Create("test_guy3"),
			Body:   "I approve version:0.0.1",
		},
	}
	request := &model.ApprovalRequest{
		Config:           config,
		Maintainer:       m,
		PullRequest:      &i,
		ApprovalComments: comments,
	}
	ver := getLastVersionComment(request, policy)
	if ver == "" {
		t.Fatal("Got empty for version")
	}
	if ver != "0.0.1" {
		t.Errorf("Expected 0.0.1, got %s", ver)
	}
}
