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
package set

import (
	"encoding/json"
	"sort"
	"github.com/capitalone/checks-out/strings/lowercase"
	"strings"
)

// LowerSet is a collection of unique lowercase strings
type LowerSet map[lowercase.String]bool

// Empty creates an empty lowerset
func EmptyLower() LowerSet {
	return make(map[lowercase.String]bool)
}

// AddAll inserts all elements from a lowerset into the lowerset
func (s LowerSet) AddAll(input LowerSet) {
	for k := range input {
		s.Add(k)
	}
}

// AddAll inserts all elements from a set into the lowerset
func (s LowerSet) AddAllSet(input Set) {
	for k := range input {
		s.Add(lowercase.Create(k))
	}
}

// Add inserts an element into the lowerset
func (s LowerSet) Add(key lowercase.String) {
	s[key] = true
}

func (s LowerSet) AddString(key string) {
	s[lowercase.Create(key)] = true
}

// Remove deletes an element from the lowerset
func (s LowerSet) Remove(key lowercase.String) {
	delete(s, key)
}

func (s LowerSet) RemoveString(key string) {
	delete(s, lowercase.Create(key))
}

// Contains tests whether an element is a member of the lowerset
func (s LowerSet) Contains(key lowercase.String) bool {
	_, ok := s[key]
	return ok
}

// ContainsString tests whether an element is a member of the lowerset
func (s LowerSet) ContainsString(key string) bool {
	_, ok := s[lowercase.Create(key)]
	return ok
}

// Intersection returns all elements in common between this lowerset and the lowerset passed in
func (s LowerSet) Intersection(other LowerSet) LowerSet {
	var small, large LowerSet
	if len(s) < len(other) {
		small, large = s, other
	} else {
		small, large = other, s
	}
	res := EmptyLower()
	for k := range small {
		if large.Contains(k) {
			res.Add(k)
		}
	}
	return res
}

// Difference returns all elements that are in this lowerset but not in the passed-in lowerset
func (s LowerSet) Difference(other LowerSet) LowerSet {
	res := EmptyLower()
	for k := range s {
		if !other.Contains(k) {
			res.Add(k)
		}
	}
	return res
}

func (s LowerSet) Print(sep string) string {
	keys := s.KeysSorted(func(s1,s2 lowercase.String) bool {
		return s1.String() < s2.String()
	})
	return strings.Join(keys.ToStringSlice(), sep)
}

// NewLower creates a new lowerset with the provided values
func NewLower(keys ...lowercase.String) LowerSet {
	set := EmptyLower()
	for _, k := range keys {
		set.Add(k)
	}
	return set
}

// NewLowerFromString creates a new lowerset with the provided values
func NewLowerFromString(keys ...string) LowerSet {
	set := EmptyLower()
	for _, k := range keys {
		set.Add(lowercase.Create(k))
	}
	return set
}

func (s LowerSet) Keys() lowercase.Slice {
	l := len(s)
	if l == 0 {
		return nil
	}
	lst := make(lowercase.Slice, 0, l)
	for k := range s {
		lst = append(lst, k)
	}
	return lst
}

func (s LowerSet) KeysSorted(f func (s1, s2 lowercase.String) bool) lowercase.Slice {
	lst := s.Keys()
	sort.Slice(lst, func(i, j int) bool {
		return f(lst[i], lst[j])
	})
	return lst
}

func (s LowerSet) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k.String())
	}
	return json.Marshal(keys)
}

func (s *LowerSet) UnmarshalJSON(data []byte) error {
	var keys []string
	err := json.Unmarshal(data, &keys)
	if err != nil {
		return err
	}
	*s = EmptyLower()
	for _, k := range keys {
		s.AddString(k)
	}
	return nil
}

func (s LowerSet) ToSet() Set {
	return New(s.Keys().ToStringSlice()...)
}