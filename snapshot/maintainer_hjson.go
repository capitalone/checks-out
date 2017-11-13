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
	"errors"
	"fmt"

	"github.com/capitalone/checks-out/hjson"
	"github.com/capitalone/checks-out/model"

	"github.com/mspiegel/go-multierror"
)

func parseMaintainerHJSON(data []byte) (*model.Maintainer, error) {
	m := new(model.Maintainer)
	err := hjson.Unmarshal(data, m)
	if err != nil {
		return nil, badRequest(err)
	}
	if m.RawPeople == nil {
		err = errors.New("Invalid HJSON format. Missing people section.")
		return nil, badRequest(err)
	}
	err = validateMaintainerHJSON(m)
	return m, err
}

func validateMaintainerHJSON(m *model.Maintainer) error {
	var errs error
	for k, v := range m.RawPeople {
		if len(v.Login) == 0 {
			// Populate the login field if it is missing.
			v.Login = k
		} else if v.Login != k {
			err := fmt.Errorf("Mismatched key %s and login field %s", k, v.Login)
			errs = multierror.Append(errs, badRequest(err))
		}
	}
	for k := range m.RawOrg {
		if ReservedOrgs.Contains(k) {
			err := fmt.Errorf("The organization name %s is a reserved name", k)
			errs = multierror.Append(errs, badRequest(err))
		}
	}
	return errs
}
