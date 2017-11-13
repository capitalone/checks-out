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
	"context"

	"github.com/capitalone/checks-out/notifier"
)

func doApprovalHook(c context.Context, hook *ApprovalHook) (*ApprovalOutput, error) {
	params, err := GetHookParameters(c, hook.Repo.Slug, true)
	if err != nil {
		return nil, err
	}
	approvalInfo, err := approve(c, params, hook.Issue.Number, true)

	if err != nil {
		notifier.SendErrorMessage(c, params.Config, hook.Issue.Title,
			hook.Issue.Number, hook.Repo.Slug, err.Error())
		return nil, err
	}
	mw := handleApprovalNotification(hook, &approvalInfo.CurCommentInfo)
	notifier.SendMessage(c, params.Config, *mw)
	approvalOutput := ApprovalOutput{
		Policy:       approvalInfo.Policy,
		Settings:     params.Config,
		Approved:     approvalInfo.Approved,
		Approvers:    approvalInfo.Approvers,
		Disapprovers: approvalInfo.Disapprovers,
	}
	return &approvalOutput, nil
}
