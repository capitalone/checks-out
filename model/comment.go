/*

SPDX-Copyright: Copyright (c) Brad Rydzewski, project contributors, Capital One Services, LLC
SPDX-License-Identifier: Apache-2.0
Copyright 2017 Brad Rydzewski, project contributors, Capital One Services, LLC

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
	"fmt"
	"regexp"
	"time"

	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/strings/lowercase"
)

var CommentPrefix = fmt.Sprintf("Message from %s --", envvars.Env.Branding.ShortName)

type Comment struct {
	Author      lowercase.String
	Body        string
	SubmittedAt time.Time
}

// IsApproval returns true if the comment body matches the regular
// expression pattern.
func (c *Comment) IsApproval(req *ApprovalRequest) bool {
	var regExp *regexp.Regexp
	policy := FindApprovalPolicy(req)
	if policy.Pattern != nil {
		regExp = policy.Pattern.Regex
	} else {
		regExp = req.Config.Pattern.Regex
	}
	if regExp == nil {
		// this should never happen
		return false
	}
	return regExp.MatchString(c.Body)
}

// IsDisapproval returns true if the comment body matches the
// antipattern regular expression.
func (c *Comment) IsDisapproval(req *ApprovalRequest) bool {
	var regExp *regexp.Regexp
	policy := FindApprovalPolicy(req)
	if policy.AntiPattern != nil {
		regExp = policy.AntiPattern.Regex
	} else if req.Config.AntiPattern != nil {
		regExp = req.Config.AntiPattern.Regex
	}
	if regExp == nil {
		// disapproval matching is optional
		return false
	}
	return regExp.MatchString(c.Body)
}

func (c *Comment) GetAuthor() lowercase.String {
	return c.Author
}

func (c *Comment) GetBody() string {
	return c.Body
}

func (c *Comment) GetSubmittedAt() time.Time {
	return c.SubmittedAt
}

func (c *Comment) String() string {
	if c == nil {
		return "nil"
	}
	return fmt.Sprintf("{Author: %s, Comment: %s}", c.Author, c.Body)
}
