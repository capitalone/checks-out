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
	"reflect"
	"testing"
)

func TestAddContains(t *testing.T) {
	s := Empty()
	s.Add("foo")
	s.Add("bar")
	s.Add("foo")
	if !s.Contains("foo") {
		t.Error("Set is missing value 'foo'")
	}
	if !s.Contains("bar") {
		t.Error("Set is missing value 'bar'")
	}
	if s.Contains("baz") {
		t.Error("Set is not missing value 'baz'")
	}
}

func TestNew(t *testing.T) {
	s := New("foo", "bar", "foo")
	if !s.Contains("foo") {
		t.Error("Set is missing value 'foo'")
	}
	if !s.Contains("bar") {
		t.Error("Set is missing value 'bar'")
	}
	if s.Contains("baz") {
		t.Error("Set is not missing value 'baz'")
	}
}

func TestKeys(t *testing.T) {
	s := New("foo", "bar", "baz")
	keys := s.Keys()
	if len(keys) != 3 {
		t.Error("Set keys list has incorrect size", keys)
	}
}

func TestPrint(t *testing.T) {
	x := New()
	y := New("a", "b")
	res1 := x.Print(",")
	res2 := y.Print(",")
	if res1 != "" {
		t.Error("Set print incorrect", res1)
	}
	if res2 != "a,b" {
		t.Error("Set print incorrect", res2)
	}
}

func TestIntersection(t *testing.T) {
	a := New("a", "b", "c")
	b := New("b", "c", "y", "z")
	obs := a.Intersection(b)
	exp := New("b", "c")
	if !reflect.DeepEqual(exp, obs) {
		t.Error("Set difference incorrect", obs)
	}
	obs = b.Intersection(a)
	if !reflect.DeepEqual(exp, obs) {
		t.Error("Set difference incorrect", obs)
	}
}

func TestJSON(t *testing.T) {
	var out Set
	in := New("foo", "bar")
	text, err := json.Marshal(in)
	if err != nil {
		t.Fatal("Error marshaling set", err)
	}
	err = json.Unmarshal(text, &out)
	if err != nil {
		t.Fatal("Error unmarshaling set", err)
	}
	if !out.Contains("foo") {
		t.Error("Set is missing value 'foo'")
	}
	if !out.Contains("bar") {
		t.Error("Set is missing value 'bar'")
	}
	if out.Contains("baz") {
		t.Error("Set is not missing value 'baz'")
	}
}
