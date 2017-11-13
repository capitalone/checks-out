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
package model

import (
	"fmt"

	"github.com/capitalone/checks-out/set"
)

// Person represets an individual in the MAINTAINERS file.
type Person struct {
	Name  string `json:"name"  toml:"name"`
	Email string `json:"email" toml:"email"`
	Login string `json:"login" toml:"login"`
}

// Org represents a group, team or subset of users.
type Org struct {
	People set.Set `json:"people"  toml:"people"`
}

// Maintainer represents a MAINTAINERS file.
type Maintainer struct {
	RawPeople map[string]*Person `json:"people" toml:"people"`
	RawOrg    map[string]*Org    `json:"org" toml:"org"`
}

var MaintTypes = set.New("text", "hjson", "toml", "legacy")

func validateMaintainerConfig(c *MaintainersConfig) error {
	if !MaintTypes.Contains(c.Type) {
		return fmt.Errorf("%s is not one of the permitted MAINTAINER types %s",
			c.Type, MaintTypes.Keys())
	}
	return nil
}
