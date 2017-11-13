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
package api

import (
	"fmt"

	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/router/middleware/session"
	"github.com/capitalone/checks-out/snapshot"
	"github.com/capitalone/checks-out/store"

	"github.com/gin-gonic/gin"
)

var configFileName = fmt.Sprintf(".%s", envvars.Env.Branding.Name)

// GetConfig gets the parsed configuration file.
func GetConfig(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("repo")
		user  = session.User(c)
		caps  = session.Capability(c)
	)
	repo, err := store.GetRepoOwnerName(c, owner, name)
	if err != nil {
		msg := fmt.Sprintf("Getting repository %s", name)
		c.Error(exterror.Append(err, msg))
		return
	}
	config, err := snapshot.GetConfig(c, user, caps, repo)
	if err != nil {
		msg := fmt.Sprintf("Getting %s configuration for %s", configFileName, name)
		c.Error(exterror.Append(err, msg))
	} else {
		IndentedJSON(c, 200, config)
	}
}
