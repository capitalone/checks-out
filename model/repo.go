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

type Repo struct {
	ID      int64  `json:"id,omitempty"       meddler:"repo_id,pk"`
	UserID  int64  `json:"-"                  meddler:"repo_user_id"`
	Owner   string `json:"owner"              meddler:"repo_owner"`
	Name    string `json:"name"               meddler:"repo_name"`
	Slug    string `json:"slug"               meddler:"repo_slug"`
	Link    string `json:"link_url"           meddler:"repo_link"`
	Private bool   `json:"private"            meddler:"repo_private"`
	Secret  string `json:"-"                  meddler:"repo_secret"`
	Org     bool   `json:"org"                meddler:"repo_org"`
}

type Perm struct {
	Pull  bool
	Push  bool
	Admin bool
}

type OrgDb struct {
	ID      int64  `json:"id,omitempty"       meddler:"org_id,pk"`
	UserID  int64  `json:"-"                  meddler:"org_user_id"`
	Owner   string `json:"owner"              meddler:"org_owner"`
	Link    string `json:"link_url"           meddler:"org_link"`
	Private bool   `json:"private"            meddler:"org_private"`
	Secret  string `json:"-"                  meddler:"org_secret"`
}
