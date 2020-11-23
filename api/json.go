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
	"bytes"
	"encoding/json"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func IndentedJSON(c *gin.Context, code int, obj interface{}) {
	var buf bytes.Buffer
	e := json.NewEncoder(&buf)
	e.SetEscapeHTML(false)
	e.SetIndent("", "    ")
	err := e.Encode(obj)
	if err != nil {
		log.Errorf("JSON encoding error %+v. %s", obj, err)
		c.String(500, "JSON encoding error")
	} else {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.String(code, string(buf.Bytes()))
	}
}
