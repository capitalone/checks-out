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
	"fmt"
	"testing"

	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/set"
)

func TestGenerateStatus(t *testing.T) {
	testCases := map[*ApprovalInfo]struct {
		status string
		desc   string
	}{
		&ApprovalInfo{
			Approved: true,
		}: {
			status: "success",
			desc:   "approval did not require approvers",
		},

		&ApprovalInfo{
			Approved:  true,
			Approvers: set.New("bob", "frank"),
		}: {
			status: "success",
			desc:   "approved by bob,frank",
		},

		&ApprovalInfo{
			Approved:       false,
			AuditApproved:  false,
			TitleApproved:  false,
			AuthorApproved: false,
			AuthorAffirmed: true,
		}: {
			status: "error",
			desc:   "audit chain must be manually approved",
		},

		&ApprovalInfo{
			Approved:       false,
			AuditApproved:  true,
			TitleApproved:  false,
			AuthorApproved: false,
			AuthorAffirmed: true,
		}: {
			status: "error",
			desc:   "pull request title is blocking merge",
		},

		&ApprovalInfo{
			Approved:       false,
			AuditApproved:  true,
			TitleApproved:  true,
			AuthorApproved: false,
			AuthorAffirmed: true,
		}: {
			status: "error",
			desc:   "pull request author not allowed",
		},

		&ApprovalInfo{
			Approved:       false,
			AuditApproved:  true,
			TitleApproved:  true,
			AuthorApproved: true,
			AuthorAffirmed: false,
		}: {
			status: "error",
			desc:   "PR author or non-commit author must approve",
		},

		&ApprovalInfo{
			Approved:       false,
			AuditApproved:  true,
			TitleApproved:  true,
			AuthorApproved: true,
			AuthorAffirmed: true,
			Disapprovers:   set.New("bob", "frank"),
		}: {
			status: "pending",
			desc:   "blocked by bob,frank",
		},

		&ApprovalInfo{
			Approved:       false,
			AuditApproved:  true,
			TitleApproved:  true,
			AuthorApproved: true,
			AuthorAffirmed: true,
			Approvers:      set.New("bob", "frank"),
		}: {
			status: "pending",
			desc:   fmt.Sprintf("more approvals needed. %s: bob,frank", envvars.Env.Branding.ShortName),
		},

		&ApprovalInfo{
			Approved:       false,
			AuditApproved:  true,
			TitleApproved:  true,
			AuthorApproved: true,
			AuthorAffirmed: true,
		}: {
			status: "pending",
			desc:   "no approvals received",
		},
	}
	for k, v := range testCases {
		status, desc := generateStatus(k)
		if v.status != status || v.desc != desc {
			t.Errorf("Expected status %s and desc %s but got %s and %s", v.status, v.desc, status, desc)
		}
	}
}
