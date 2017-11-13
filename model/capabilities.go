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
	"errors"

	"github.com/capitalone/checks-out/set"

	"github.com/mspiegel/go-multierror"
)

type Capabilities struct {
	Org struct {
		Read bool
	}
	Repo struct {
		Tag              bool
		Merge            bool
		DeleteBranch     bool
		CommitStatus     bool
		PRWriteComment   bool
		DeploymentStatus bool
	}
}

func AllowAll() *Capabilities {
	caps := new(Capabilities)
	caps.Org.Read = true
	caps.Repo.Tag = true
	caps.Repo.Merge = true
	caps.Repo.DeleteBranch = true
	caps.Repo.CommitStatus = true
	caps.Repo.PRWriteComment = true
	caps.Repo.DeploymentStatus = true
	return caps
}

func validateCapabilities(c *Config, caps *Capabilities) error {
	var errs error
	errMsgs := set.Empty()
	if !caps.Repo.CommitStatus {
		errMsgs.Add("commit status OAuth scope is required")
	}
	if c.Tag.Enable && !caps.Repo.Tag {
		errMsgs.Add("unable to git tag with provided OAuth scopes")
	}
	if c.Merge.Enable && !caps.Repo.Merge {
		errMsgs.Add("unable to git merge with provided OAuth scopes")
	}
	if c.Merge.Enable && c.Merge.Delete && !caps.Repo.DeleteBranch {
		errMsgs.Add("unable to delete branch with provided OAuth scopes")
	}
	for _, policy := range c.Approvals {
		if policy.Tag != nil && policy.Tag.Enable && !caps.Repo.Tag {
			errMsgs.Add("unable to git tag with provided OAuth scopes")
		}
		if policy.Merge != nil && policy.Merge.Enable && !caps.Repo.Merge {
			errMsgs.Add("unable to git merge with provided OAuth scopes")
		}
		if policy.Merge != nil && policy.Merge.Enable && policy.Merge.Delete && !caps.Repo.DeleteBranch {
			errMsgs.Add("unable to delete branch with provided OAuth scopes")
		}
	}
	if c.Comment.Enable {
		for _, target := range c.Comment.Targets {
			if target.Target == Github.String() && !caps.Repo.PRWriteComment {
				errMsgs.Add("unable to add PR comment with provided OAuth scopes")
			}
		}
	}
	for msg := range errMsgs {
		errs = multierror.Append(errs, errors.New(msg))
	}
	return errs
}
