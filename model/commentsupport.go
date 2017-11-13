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
)

type CommentMessage int

const (
	_ CommentMessage = iota
	//error while processing webhook
	CommentError
	//pr is opened
	CommentOpen
	//pr is closed without being merged
	CommentClose
	//pr has been closed and merged
	CommentAccept
	//an approval comment has been added to the PR
	CommentApprove
	//a blocking comment has been added to the PR
	CommentBlock
	//new push on PR branch has reset previous approvals
	CommentReset
	//a merge via the user interface has been ignored
	CommentPushIgnore
	//pr was auto-merged after all status checks passed
	CommentMerge
	//repo was tagged after merge
	CommentTag
	//branch for pr was auto-deleted after merge
	CommentDelete
	//deployment was triggered after merge
	CommentDeployment
	//pull request blocked because of author
	CommentAuthor
)

// CommentMessage enum maps.
var (
	strMapCommentMessage = map[string]CommentMessage{
		"error":       CommentError,
		"open":        CommentOpen,
		"close":       CommentClose,
		"accept":      CommentAccept,
		"approve":     CommentApprove,
		"block":       CommentBlock,
		"reset":       CommentReset,
		"push-ignore": CommentPushIgnore,
		"merge":       CommentMerge,
		"tag":         CommentTag,
		"delete":      CommentDelete,
		"deploy":      CommentDeployment,
		"author":      CommentAuthor,
	}

	intMapCommentMessage = map[CommentMessage]string{
		CommentError:      "error",
		CommentOpen:       "open",
		CommentClose:      "close",
		CommentAccept:     "accept",
		CommentApprove:    "approve",
		CommentBlock:      "block",
		CommentReset:      "reset",
		CommentPushIgnore: "push-ignore",
		CommentMerge:      "merge",
		CommentTag:        "tag",
		CommentDelete:     "delete",
		CommentDeployment: "deploy",
		CommentAuthor:     "author",
	}
)

// Known says whether or not this value is a known enum value.
func (s CommentMessage) Known() bool {
	_, ok := intMapCommentMessage[s]
	return ok
}

// String is for the standard stringer interface.
func (s CommentMessage) String() string {
	return intMapCommentMessage[s]
}

// UnmarshalJSON satisfies the json.Unmarshaler
func (s *CommentMessage) UnmarshalJSON(data []byte) error {
	str := ""
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	var ok bool
	*s, ok = strMapCommentMessage[str]
	if !ok {
		return fmt.Errorf("Unknown CommentMessage enum value: %s", str)
	}
	return nil
}

// MarshalJSON satisfies the json.Marshaler
func (s CommentMessage) MarshalJSON() ([]byte, error) {
	if !s.Known() {
		return nil, fmt.Errorf("Unknown CommentMessage enum value: %d", int(s))
	}
	name := intMapCommentMessage[s]
	return json.Marshal(name)
}

type CommentTarget int

const (
	_ CommentTarget = iota
	Github
	Slack
)

// CommentTarget enum maps.
var (
	strMapCommentTarget = map[string]CommentTarget{
		"github": Github,
		"slack":  Slack,
	}

	intMapCommentTarget = map[CommentTarget]string{
		Github: "github",
		Slack:  "slack",
	}
)

func ToCommentTarget(s string) CommentTarget {
	return strMapCommentTarget[s]
}

// Known says whether or not this value is a known enum value.
func (s CommentTarget) Known() bool {
	_, ok := intMapCommentTarget[s]
	return ok
}

// String is for the standard stringer interface.
func (s CommentTarget) String() string {
	return intMapCommentTarget[s]
}

// UnmarshalJSON satisfies the json.Unmarshaler
func (s *CommentTarget) UnmarshalJSON(data []byte) error {
	str := ""
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	var ok bool
	*s, ok = strMapCommentTarget[str]
	if !ok {
		return fmt.Errorf("Unknown CommentTarget enum value: %s", str)
	}
	return nil
}

// MarshalJSON satisfies the json.Marshaler
func (s CommentTarget) MarshalJSON() ([]byte, error) {
	if !s.Known() {
		return nil, fmt.Errorf("Unknown CommentTarget enum value: %d", int(s))
	}
	name := intMapCommentTarget[s]
	return json.Marshal(name)
}
