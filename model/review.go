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
	"time"

	"github.com/capitalone/checks-out/strings/lowercase"
)

type Review struct {
	ID          int
	Author      lowercase.String
	Body        string
	SubmittedAt time.Time
	State       lowercase.String
}

// IsApproval returns true if the review has been approved
func (r *Review) IsApproval(req *ApprovalRequest) bool {
	return r.State.String() == "approved"
}

// IsDisapproval returns true if changes have been requested
func (r *Review) IsDisapproval(req *ApprovalRequest) bool {
	return r.State.String() == "changes_requested"
}

func (r *Review) GetAuthor() lowercase.String {
	return r.Author
}

func (r *Review) GetBody() string {
	return r.Body
}

func (r *Review) GetSubmittedAt() time.Time {
	return r.SubmittedAt
}
