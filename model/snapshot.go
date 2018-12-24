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
package model

import "github.com/capitalone/checks-out/set"

type MaintainerSnapshot struct {
	People map[string]*Person
	Org    map[string]Org
}

func (m *MaintainerSnapshot) PersonToOrg() (map[string]set.Set, error) {
	mapping := make(map[string]set.Set)
	for k, v := range m.Org {
		//value is name of person in the org
		people, err := v.GetPeople()
		if err != nil {
			return nil, err
		}
		for name := range people {
			if _, ok := m.People[name]; !ok {
				continue
			}
			orgs, ok := mapping[name]
			if !ok {
				orgs = set.Empty()
				mapping[name] = orgs
			}
			orgs.Add(k)
		}
	}
	return mapping, nil
}
