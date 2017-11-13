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
package snapshot

import (
	"github.com/capitalone/checks-out/model"

	multierror "github.com/mspiegel/go-multierror"
)

func validateSnapshot(config *model.Config, snapshot *model.MaintainerSnapshot) error {
	var errs error
	for _, approval := range config.Approvals {
		err := approval.Match.Validate(snapshot)
		errs = multierror.Append(errs, err)
		if approval.AntiMatch != nil {
			err := approval.AntiMatch.Validate(snapshot)
			errs = multierror.Append(errs, err)
		}
		if approval.AuthorMatch != nil {
			err := approval.AuthorMatch.Validate(snapshot)
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}
