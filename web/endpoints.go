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
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ApprovalStatus generates a response with the current status of a pull request
func ApprovalStatus(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("repo")
		id    = c.Param("id")
	)

	pr, err := strconv.Atoi(id)
	if err != nil {
		c.String(400, "Unable to convert pull request id %s to number", id)
		return
	}

	params, err := GetHookParameters(c, path.Join(owner, name), false)
	if err != nil {
		c.Error(err)
		return
	}
	approvalInfo, err := approve(c, params, pr, false)
	if err != nil {
		c.Error(err)
	} else {
		c.IndentedJSON(200, gin.H{
			"policy":       approvalInfo.Policy,
			"settings":     params.Config,
			"approved":     approvalInfo.Approved,
			"approvers":    approvalInfo.Approvers,
			"disapprovers": approvalInfo.Disapprovers,
		})
	}
}
