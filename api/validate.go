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
	"context"
	"fmt"

	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/hjson"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/router/middleware/session"
	"github.com/capitalone/checks-out/snapshot"
	"github.com/capitalone/checks-out/store"

	"github.com/gin-gonic/gin"
)

const (
	updateMessage = `You are still using a .lgtm file to configure your project.
	 If you want to take advantage of the new features in Checks-Out, you will need to upgrade to the new file format first.
	 To upgrade, create a configuration file with the following content and check it into your default branch: `
)

var okMessage = fmt.Sprintf(`Your %s and MAINTAINERS files are valid.`, configFileName)

func validateRepo(c context.Context, repo *model.Repo, owner string, name string, user *model.User, caps *model.Capabilities) (string, error) {
	config, _, err := snapshot.GetConfigAndMaintainers(c, user, caps, repo)
	if err != nil {
		msg := fmt.Sprintf("Error validating %s", name)
		return "", exterror.Append(err, msg)
	}
	err = snapshot.FixSlackTargets(c, config, user.Login)
	if err != nil {
		msg := fmt.Sprintf("Error validating %s", name)
		return "", exterror.Append(err, msg)
	}
	if config.IsOld {
		//this is a bit of a hack, but we're going to strip out the bits of the config that
		//were added to the contents present in the configuration file
		config.Approvals[0].Scope = nil
		config.Approvals[0].AntiMatch = nil
		config.Approvals[0].AuthorMatch = nil
		config.Maintainers = model.MaintainersConfig{}
		options := hjson.DefaultOptions()
		options.OmitEmptyStructs = true
		out, err := hjson.MarshalWithOptions(config, options)
		if err != nil {
			msg := fmt.Sprintf("Error converting .lgtm file for %s", name)
			return "", exterror.Append(err, msg)
		}
		return string(out), nil
	}
	return "", nil
}

func validateAndGetFile(c context.Context, owner string, name string, user *model.User, caps *model.Capabilities) (string, error) {
	repo, err := store.GetRepoOwnerName(c, owner, name)
	if err != nil {
		msg := fmt.Sprintf("Getting repository %s", name)
		return "", exterror.Append(err, msg)

	}
	return validateRepo(c, repo, owner, name, user, caps)
}

// Validate validates the configuration and MAINTAINER files.
func Validate(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("repo")
		user  = session.User(c)
		caps  = session.Capability(c)
	)
	file, err := validateAndGetFile(c, owner, name, user, caps)
	if err != nil {
		c.Error(err)
		return
	}
	message := okMessage
	if file != "" {
		message = updateMessage
	}
	c.JSON(200, map[string]string{
		"message": message,
		"file":    file,
	})
}

func Convert(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("repo")
		user  = session.User(c)
		caps  = session.Capability(c)
	)
	file, err := validateAndGetFile(c, owner, name, user, caps)
	if err != nil {
		c.Error(err)
		return
	}
	status := 200
	if len(file) == 0 {
		status = 204
	}
	c.String(status, file)
}
