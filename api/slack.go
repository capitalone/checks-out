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
	"encoding/json"
	"fmt"
	"github.com/capitalone/checks-out/exterror"
	"github.com/capitalone/checks-out/router/middleware/session"
	"github.com/capitalone/checks-out/store"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

type UrlHolder struct {
	Url string `json:"url"`
}

func RegisterSlackUrl(c *gin.Context) {
	registerSlackUrlInner(c, "", 201)
}

func registerSlackUrlInner(c *gin.Context, user string, successCode int) {
	var (
		hostname = c.Param("hostname")
	)
	var urlHolder UrlHolder
	defer func() {
		c.Request.Body.Close()
	}()
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		msg := fmt.Sprintf("Registering slack URL by user %s for host %s ", user, hostname)
		c.Error(exterror.Append(err, msg))
		c.Error(err)
	}
	err = json.Unmarshal(body, &urlHolder)
	if err != nil {
		msg := fmt.Sprintf("Registering slack URL by user %s for host %s ", user, hostname)
		c.Error(exterror.Append(err, msg))
		c.Error(err)
	}

	err = store.AddUpdateSlackUrl(c, hostname, user, urlHolder.Url)
	if err != nil {
		msg := fmt.Sprintf("Registering slack URL %s by user %s for host %s ", urlHolder.Url, user, hostname)
		c.Error(exterror.Append(err, msg))
		c.Error(err)
	} else {
		c.Status(successCode)
	}
}

func getSlackUrlInner(c *gin.Context, user string) {
	var (
		hostname = c.Param("hostname")
	)

	url, err := store.GetSlackUrl(c, hostname, user)
	if err != nil {
		msg := fmt.Sprintf("Getting slack URL by user %s for host %s ", user, hostname)
		c.Error(exterror.Append(err, msg))
		c.Error(err)
	} else {
		if url == "" {
			c.Status(404)
		} else {
			c.IndentedJSON(200, UrlHolder{
				Url: url,
			})
		}
	}
}

func deleteSlackUrlInner(c *gin.Context, user string) {
	var (
		hostname = c.Param("hostname")
	)

	err := store.DeleteSlackUrl(c, hostname, user)
	if err != nil {
		msg := fmt.Sprintf("Deleting slack URL by user %s for host %s ", user, hostname)
		c.Error(exterror.Append(err, msg))
		c.Error(err)
	} else {
		c.Status(204)
	}
}

func GetSlackUrl(c *gin.Context) {
	getSlackUrlInner(c, "")
}

func DeleteSlackUrl(c *gin.Context) {
	deleteSlackUrlInner(c, "")
}

func UpdateSlackUrl(c *gin.Context) {
	registerSlackUrlInner(c, "", 200)
}

func UserRegisterSlackUrl(c *gin.Context) {
	user := session.User(c)
	registerSlackUrlInner(c, user.Login, 201)
}

func UserGetSlackUrl(c *gin.Context) {
	user := session.User(c)
	getSlackUrlInner(c, user.Login)

}

func UserDeleteSlackUrl(c *gin.Context) {
	user := session.User(c)
	deleteSlackUrlInner(c, user.Login)
}

func UserUpdateSlackUrl(c *gin.Context) {
	user := session.User(c)
	registerSlackUrlInner(c, user.Login, 200)
}
