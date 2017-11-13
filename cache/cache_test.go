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
package cache

import (
	"testing"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"
)

func TestCache(t *testing.T) {

	g := goblin.Goblin(t)
	g.Describe("Cache", func() {

		var c *gin.Context
		g.BeforeEach(func() {
			c = new(gin.Context)
			ToContext(c, Default())
		})

		g.It("Should set and get an item", func() {
			Set(c, "foo", "bar")
			v, e := Get(c, "foo")
			g.Assert(v).Equal("bar")
			g.Assert(e == nil).IsTrue()
		})

		g.It("Should return nil when item not found", func() {
			v, e := Get(c, "foo")
			g.Assert(v == nil).IsTrue()
			g.Assert(e == nil).IsFalse()
		})
	})
}
