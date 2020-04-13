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
package web

import (
	"fmt"

	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/notifier"

	log "github.com/sirupsen/logrus"
)

func handleApprovalNotification(hook *ApprovalHook, curCommentInfo *CurCommentInfo) *notifier.MessageWrapper {
	mw := &notifier.MessageWrapper{
		MessageHeader: notifier.MessageHeader{
			PrName:   hook.Issue.Title,
			PrNumber: hook.Issue.Number,
			Slug:     hook.Repo.Slug,
		},
	}
	if curCommentInfo != nil {
		var mi notifier.MessageInfo
		switch curCommentInfo.Status {
		case CurCommentNoChange:
			// do nothing
		case CurCommentApproval:
			mi.Message = fmt.Sprintf("approval added by %s.", curCommentInfo.Author)
			mi.Type = model.CommentApprove
		case CurCommentDisapproval:
			mi.Message = fmt.Sprintf("blocked by %s.", curCommentInfo.Author)
			mi.Type = model.CommentBlock
		case CurCommentPRAuthor:
			mi.Message = fmt.Sprintf("blocked because it was created by unapproved author %s.", hook.Issue.Author)
			mi.Type = model.CommentBlock
		case CurCommentPRTitle:
			mi.Message = "blocked because its title indicates that it should not be merged"
			mi.Type = model.CommentBlock
		case CurCommentPRAudit:
			mi.Message = "blocked by gap in audit chain"
			mi.Type = model.CommentBlock
		default:
			log.Warnf("Invalid curCommentInfo.Status found, skipping: %v", curCommentInfo.Status)
		}

		if mi.Message != "" {
			mw.Messages = append(mw.Messages, mi)
		}
	}
	return mw
}
