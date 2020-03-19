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
package middleware

import (
	"fmt"

	"github.com/capitalone/checks-out/exterror"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ExtError() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		errs := c.Errors
		if len(errs) == 1 {
			logAndRespond(c, errs[0].Err)
		} else if len(errs) > 1 {
			err := fmt.Errorf("Multiple errors: %s", errs.String())
			err = exterror.ExtError{Status: 500, Err: err}
			logAndRespond(c, err)
		}
	}
}

func emitLog(c *gin.Context, e exterror.ExtError) {
	msg := e.Err.Error()
	if e.Status < 500 {
		log.Warn(msg)
	} else {
		log.Error(msg)
		for k, v := range c.Request.Header {
			log.Errorf("%s: %v", k, v)
		}
	}
}

func logAndRespond(c *gin.Context, e error) {
	err := exterror.Convert(e)
	emitLog(c, err)
	c.String(err.Status, err.Error())
}
