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
package lowercase

import (
	"strings"
)

type String struct {
	val string
}

func Create(s string) String {
	return String{strings.ToLower(s)}
}

func (l String) String() string {
	return l.val
}

type Slice []String

func CreateSlice(ss ...string) Slice {
	var out Slice
	for _, v := range ss {
		out = append(out, Create(v))
	}
	return out
}

func (ls Slice) ToStringSlice() []string {
	var out []string
	for _, v := range ls {
		out = append(out, v.String())
	}
	return out
}
