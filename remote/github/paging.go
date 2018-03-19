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
package github

import (
	"github.com/mspiegel/go-github/github"
)

func buildCompleteList(process func(opts *github.ListOptions) (*github.Response, error)) (*github.Response, error) {
	var response *github.Response
	var err error
	opts := &github.ListOptions{}
	for {
		response, err = process(opts)
		if err != nil || response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}
	return response, err
}
