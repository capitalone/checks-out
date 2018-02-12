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

	multierror "github.com/mspiegel/go-multierror"
)

func (hook *ReviewHook) Process(c context.Context) (interface{}, error) {
	approvalOutput, e1 := doApprovalHook(c, &hook.ApprovalHook, hook)
	if e1 != nil {
		e2 := sendErrorStatus(c, &hook.ApprovalHook, e1)
		e1 = multierror.Append(e1, e2)
	}
	return approvalOutput, e1
}
