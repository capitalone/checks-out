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
)

// Set is a collection of unique strings
type Set map[string]bool

// Empty creates an empty set
func Empty() Set {
	return make(map[string]bool)
}

// AddAll inserts all element into the set
func (s Set) AddAll(input Set) {
	for k := range input {
		s.Add(k)
	}
}

// Add inserts an element into the set
func (s Set) Add(key string) {
	s[key] = true
}

// Remove deletes an element from the set
func (s Set) Remove(key string) {
	delete(s, key)
}

// Contains tests whether an element is a member of the set
func (s Set) Contains(key string) bool {
	_, ok := s[key]
	return ok
}

func (s Set) Intersection(other Set) Set {
	var small, large Set
	if len(s) < len(other) {
		small, large = s, other
	} else {
		small, large = other, s
	}
	res := Empty()
	for k := range small {
		if large.Contains(k) {
			res.Add(k)
		}
	}
	return res
}

func (s Set) Difference(other Set) Set {
	res := Empty()
	for k := range s {
		if !other.Contains(k) {
			res.Add(k)
		}
	}
	return res
}

func (s Set) Print(sep string) string {
	res := ""
	keys := s.KeysSorted(func(s1, s2 string) bool {
		return s1 < s2
	})
	for i, k := range keys {
		res += k
		if i < len(keys)-1 {
			res += sep
		}
	}
	return res
}

// New creates a new set with the provided values
func New(keys ...string) Set {
	set := Empty()
	for _, k := range keys {
		set.Add(k)
	}
	return set
}

func (s Set) Keys() []string {
	l := len(s)
	if l == 0 {
		return nil
	}
	lst := make([]string, 0, l)
	for k := range s {
		lst = append(lst, k)
	}
	return lst
}

func (s Set) KeysSorted(f func(s1, s2 string) bool) []string {
	lst := s.Keys()
	sort.Slice(lst, func(i, j int) bool {
		return f(lst[i], lst[j])
	})
	return lst
}

func (s Set) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return json.Marshal(keys)
}

func (s *Set) UnmarshalJSON(data []byte) error {
	var keys []string
	err := json.Unmarshal(data, &keys)
	if err != nil {
		return err
	}
	*s = Empty()
	for _, k := range keys {
		s.Add(k)
	}
	return nil
}
