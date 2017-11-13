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

	"github.com/capitalone/checks-out/model"
	"github.com/capitalone/checks-out/set"

	"github.com/pelletier/go-toml"
)

func parseMaintainerToml(data []byte) (*model.Maintainer, error) {
	tree, err := toml.Load(string(data))
	if err != nil {
		return nil, badRequest(err)
	}

	m := new(model.Maintainer)

	if tree.Has("people") {
		m.RawPeople, err = buildRawPeople(tree)
		if err != nil {
			return nil, badRequest(err)
		}
	}

	if tree.Has("org") {
		m.RawOrg, err = buildRawOrg(tree)
		if err != nil {
			return nil, badRequest(err)
		}
	}

	if m.RawPeople == nil {
		err = errors.New("Invalid Toml format. Missing people section.")
		return nil, badRequest(err)
	}

	err = validateMaintainerHJSON(m)
	return m, err
}

func buildRawOrg(tree *toml.Tree) (map[string]*model.Org, error) {
	rawOrg := map[string]*model.Org{}
	o, ok := tree.Get("org").(*toml.Tree)
	if !ok {
		return nil, errors.New("Invalid Toml format. Org section is invalid.")
	}
	for _, v := range o.Keys() {
		org, ok := o.Get(v).(*toml.Tree)
		if !ok {
			return nil, fmt.Errorf("Invalid toml format. Org %s is invalid.", v)
		}
		curOrg := &model.Org{}
		pSlice, ok := org.Get("people").([]interface{})
		if !ok {
			return nil, fmt.Errorf("Invalid toml format. Org %s people isn't a slice of strings", v)
		}
		curOrg.People = set.Empty()
		for _, v2 := range pSlice {
			if p, ok := v2.(string); ok {
				curOrg.People.Add(p)
			} else {
				return nil, fmt.Errorf("Invalid toml format. Org %s people had invalid person %v", v, v2)
			}
		}
		rawOrg[v] = curOrg
	}
	return rawOrg, nil
}

func buildRawPeople(tree *toml.Tree) (map[string]*model.Person, error) {
	rawPeople := map[string]*model.Person{}
	p, ok := tree.Get("people").(*toml.Tree)
	if !ok {
		return nil, errors.New("Invalid Toml format. People section is invalid.")
	}
	for _, k := range p.Keys() {
		person, ok := p.Get(k).(*toml.Tree)
		if !ok {
			return nil, fmt.Errorf("Invalid toml format. Person %s is invalid.", k)
		}
		curPerson := &model.Person{}
		// Populate the login field if it is missing.
		str, err := extractString("login", k, person)
		if err != nil {
			return nil, err
		}
		curPerson.Login = str

		str, err = extractString("name", "", person)
		if err != nil {
			return nil, err
		}
		curPerson.Name = str

		str, err = extractString("email", "", person)
		if err != nil {
			return nil, err
		}
		curPerson.Email = str

		rawPeople[k] = curPerson
	}
	return rawPeople, nil
}

func extractString(key string, def string, person *toml.Tree) (string, error) {
	if str, ok := person.GetDefault(key, def).(string); !ok {
		return "", fmt.Errorf("Invalid toml format. Invalid field %s", key)
	} else {
		return str, nil
	}
}
