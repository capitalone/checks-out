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
	"strings"
	"time"

	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/exterror"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Ginrus(logger *logrus.Logger, timeFormat string, utc bool) gin.HandlerFunc {
	uaslice := strings.Split(envvars.Env.Monitor.UaList, ":")
	if len(uaslice) == 1 && len(uaslice[0]) == 0 {
		uaslice = nil
	}
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		c.Next()

		userAgent := c.Request.UserAgent()

		for _, v := range uaslice {
			if strings.Contains(userAgent, v) {
				return
			}
		}
		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		entry := logger.WithFields(logrus.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user-agent": userAgent,
			"time":       end.Format(timeFormat),
		})

		if len(c.Errors) > 1 {
			// Append error field if this is an erroneous request.
			entry.Error(c.Errors.String())
		} else if len(c.Errors) == 1 {
			err := c.Errors[0].Err
			if exterror.Convert(err).Status < 500 {
				entry.Warn(err.Error())
			} else {
				entry.Error(err.Error())
			}
		} else {
			entry.Info()
		}
	}
}
