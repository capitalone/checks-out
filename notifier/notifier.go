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
package notifier

import (
	"context"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/capitalone/checks-out/model"
)

var (
	senders = map[model.CommentTarget]Sender{}
)

type Sender interface {
	Send(c context.Context, header MessageHeader, message string, names []string, url string)
	Prefix(mw MessageWrapper) string
}

type MessageWrapper struct {
	MessageHeader
	Messages []MessageInfo
}

func (mw *MessageWrapper) Merge(in MessageWrapper) {
	if mw.PrNumber != in.PrNumber || mw.PrName != in.PrName {
		log.Warnf("Attempted to merge together incompatible message wrappers -- shouldn't ever happen, skipping: %v, %v", *mw, in)
		return
	}
	mw.Messages = append(mw.Messages, in.Messages...)
}

type MessageHeader struct {
	PrName   string
	PrNumber int
	Slug     string
}

type MessageInfo struct {
	Message string
	Type    model.CommentMessage
}

// Register is called by the init methods of the notification
// systems at startup. If we have something bad for registering
// messaging endpoints then we should not start the service up.
func Register(name model.CommentTarget, sender Sender) {
	if _, ok := senders[name]; ok {
		panic("Duplicate Sender name " + name.String())
	}
	senders[name] = sender
}

func BuildErrorMessage(prName string, prNumber int, slug string, message string) MessageWrapper {
	return MessageWrapper{
		MessageHeader: MessageHeader{
			PrName:   prName,
			PrNumber: prNumber,
			Slug:     slug,
		},
		Messages: []MessageInfo{
			{
				Message: message,
				Type:    model.CommentError,
			},
		},
	}
}

func SendErrorMessage(c context.Context, config *model.Config, prName string, prNumber int, slug string, message string) {
	SendMessage(c, config, BuildErrorMessage(prName, prNumber, slug, message))
}

func hasMessageType(mi MessageInfo, tc model.TargetConfig) bool {
	if tc.Types == nil {
		return true
	}
	for _, curType := range tc.Types {
		if curType == mi.Type {
			return true
		}
	}
	return false
}

func SendMessage(c context.Context, config *model.Config, mw MessageWrapper) {
	if !config.Comment.Enable {
		return
	}
	for _, v := range config.Comment.Targets {
		sender, ok := senders[model.ToCommentTarget(v.Target)]
		if !ok {
			log.Warnf("Unregistered sender %s; skipping", v.Target)
			continue
		}
		//check if the filter applies
		if v.Pattern != nil && !v.Pattern.Regex.MatchString(mw.PrName) {
			continue
		}
		var messages []string
		for _, mi := range mw.Messages {
			//check if the message type matches
			if hasMessageType(mi, v) {
				messages = append(messages, mi.Message)
			}
		}
		if len(messages) == 0 {
			continue
		}
		message := strings.Join(messages, "\n")
		message = fmt.Sprintf("%s%s", sender.Prefix(mw), message)
		sender.Send(c, mw.MessageHeader, message, v.Names, v.Url)
	}
}
