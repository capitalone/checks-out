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
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/capitalone/checks-out/envvars"
	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/notifier"

	log "github.com/sirupsen/logrus"
)

var (
	githubUrl  = envvars.Env.Github.Url
	httpClient = &http.Client{}
)

func init() {
	notifier.Register(model.Slack, &MySender{})
}

type MySender struct{}

//https://github.com/capitalone/checks-out/pull/205
func (ms *MySender) Prefix(mw notifier.MessageWrapper) string {
	// <https://alert-system.com/alerts/1234|Click here>
	return fmt.Sprintf("<%s/%s/pull/%d|Pull Request %s> in <%s/%s|repo %s>: ",
		githubUrl,
		mw.MessageHeader.Slug,
		mw.MessageHeader.PrNumber,
		mw.MessageHeader.PrName,
		githubUrl,
		mw.MessageHeader.Slug,
		mw.MessageHeader.Slug)
}

func (ms *MySender) Send(c context.Context, header notifier.MessageHeader, message string, names []string, url string) {
	if url == "" {
		log.Warn("Error sending to Slack: SLACK_TARGET_URL is not configured")
		return
	}

	baseURL := c.Value("BASE_URL")
	iconURL := ""
	if baseURL != nil {
		iconURL = baseURL.(string) + "/static/images/meowser.png"
	}

	d := struct {
		Channel  string `json:"channel"`
		Text     string `json:"text"`
		Username string `json:"username"`
		IconURL  string `json:"icon_url"`
	}{
		Text:     message,
		Username: "Meowser",
		IconURL:  iconURL,
	}
	for _, name := range names {
		d.Channel = name
		output, err := toJson(d)
		if err != nil {
			log.Warnf("Unable to convert notification to JSON: %v", err)
			continue
		}
		go func(msg string) {
			//write to the slack channel
			_, err := httpClient.Post(url, "application/json", strings.NewReader(msg))
			if err != nil {
				log.Warnf("Error while writing %s to %s: %v", msg, url, err)
			}
		}(output)
	}
}

func toJson(s interface{}) (string, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	err := e.Encode(s)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
