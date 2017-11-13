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
package github

import (
	"context"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/notifier"
	"github.com/capitalone/checks-out/remote"
	"github.com/capitalone/checks-out/web"
)

type MySender struct{}

func (ms *MySender) Prefix(mw notifier.MessageWrapper) string {
	return fmt.Sprintf("Pull Request %s in repo %s: ", mw.MessageHeader.PrName, mw.MessageHeader.Slug)
}

func (ms *MySender) Send(c context.Context, header notifier.MessageHeader, message string, names []string, url string) {
	if header.PrNumber <= 0 {
		return
	}
	repo, user, caps, err := web.GetRepoAndUser(c, header.Slug)
	if err != nil {
		log.Warnf("Error retrieving GitHub information: %v", err)
		return
	}
	if !caps.Repo.PRWriteComment {
		return
	}

	//ignore names
	err = remote.WriteComment(c, user, repo, header.PrNumber, message)
	if err != nil {
		log.Warnf("Error sending GitHub notification: %v", err)
	}
}

func init() {
	notifier.Register(model.Github, &MySender{})
}
