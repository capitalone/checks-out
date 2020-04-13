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
package session

import (
	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/remote"
	"github.com/gin-gonic/gin"
)

func Capability(c *gin.Context) *model.Capabilities {
	v, ok := c.Get("capabilities")
	if !ok {
		return nil
	}
	u, ok := v.(*model.Capabilities)
	if !ok {
		return nil
	}
	return u
}

func SetCapability(c *gin.Context) {
	user := User(c)
	if user != nil {
		caps, err := remote.Capabilities(c, user)
		if err != nil {
			err2 := exterror.Convert(err)
			c.String(err2.Status, err2.Err.Error())
			c.Abort()
			return
		}
		c.Set("capabilities", caps)
	}
	c.Next()
}
