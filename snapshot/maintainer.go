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
	"context"
	"fmt"

	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/set"
)

const SelfRepo = "repo-self"
const SelfTeam = "repo-self"

var ReservedOrgs = set.New("all", "us", "them", "universe")

// ParseMaintainer parses a projects MAINTAINERS file and returns
// the list of maintainers.
func ParseMaintainer(c context.Context, user *model.User, data []byte, r *model.Repo, typ string) (*model.Maintainer, error) {
	switch typ {
	case "text":
		return parseMaintainerText(c, user, data, r)
	case "hjson":
		return parseMaintainerHJSON(data)
	case "toml":
		return parseMaintainerToml(data)
	case "legacy":
		//try to do toml, then do text -- only for .lgtm files
		m, err := parseMaintainerToml(data)
		if err != nil {
			m, err = parseMaintainerText(c, user, data, r)
		}
		return m, err

	default:
		err := fmt.Errorf("%s is not one of the permitted MAINTAINER types %s",
			typ, model.MaintTypes.Keys())
		return nil, badRequest(err)
	}
}
