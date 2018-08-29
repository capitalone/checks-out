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
package access

import (
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/router/middleware/session"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func OwnerAdmin(c *gin.Context) {
	var (
		owner = c.Param("owner")
		user  = session.User(c)
	)

	perm, err := remote.GetOrgPerm(c, user, owner)
	if err != nil {
		log.Warnf("Cannot find org %s. %s", owner, err)
		c.String(404, "Not Found")
		c.Abort()
		return
	}
	if !perm.Admin {
		log.Warnf("User %s does not have Admin access to org %s", user.Login, owner)
		c.String(403, "Insufficient privileges")
		c.Abort()
		return
	}
	log.Debugf("User %s granted Admin access to org %s", user.Login, owner)
	c.Next()
}

func RepoAdmin(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("repo")
		user  = session.User(c)
	)

	perm, err := remote.GetPerm(c, user, owner, name)
	if err != nil {
		log.Warnf("Cannot find repository %s/%s. %s", owner, name, err)
		c.String(404, "Not Found")
		c.Abort()
		return
	}
	if !perm.Admin {
		log.Warnf("User %s does not have Admin access to repository %s/%s", user.Login, owner, name)
		c.String(403, "Insufficient privileges")
		c.Abort()
		return
	}
	log.Debugf("User %s granted Admin access to %s/%s", user.Login, owner, name)
	c.Next()
}

func RepoPull(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("repo")
		user  = session.User(c)
	)

	perm, err := remote.GetPerm(c, user, owner, name)
	if err != nil {
		log.Warnf("Cannot find repository %s/%s. %s", owner, name, err)
		c.String(404, "Not Found")
		c.Abort()
		return
	}
	if !perm.Pull {
		log.Warnf("User %s does not have Pull access to repository %s/%s", user.Login, owner, name)
		c.String(404, "Not Found")
		c.Abort()
		return
	}
	log.Debugf("User %s granted Pull access to %s/%s", user.Login, owner, name)
	c.Next()
}
