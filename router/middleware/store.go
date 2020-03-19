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
package middleware

import (
	"fmt"

	"github.com/capitalone/checks-out/store"
	"github.com/capitalone/checks-out/store/datastore"
	"github.com/gin-gonic/gin"
)

func Store() gin.HandlerFunc {
	return func(c *gin.Context) {
		store.ToContext(c, datastore.Get())
		c.Next()
	}
}

func BaseURL(c *gin.Context) {
	url := c.Request.URL
	c.Set("BASE_URL", fmt.Sprintf("%s://%s", url.Scheme, url.Host))
	c.Next()
}
