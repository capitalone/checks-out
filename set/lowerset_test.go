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
	"github.com/capitalone/checks-out/strings/lowercase"
)

func TestAddAllLower(t *testing.T) {
	s := EmptyLower()
	s.AddAll(NewLowerFromString("A", "B", "C"))
	if !s.Contains(lowercase.Create("a")) {
		t.Error("Set is missing value 'a'")
	}
	if !s.Contains(lowercase.Create("b")) {
		t.Error("Set is missing value 'b'")
	}
	// We pass in upper-case, it matches lowercase
	if !s.ContainsString("C") {
		t.Error("Set is missing value 'C'")
	}
	if !s.ContainsString("c") {
		t.Error("Set is missing value 'c'")
	}
}

func TestAddContainsLower(t *testing.T) {
	s := EmptyLower()
	s.AddString("Foo")
	s.AddString("Bar")
	s.AddString("foo")
	if !s.ContainsString("foo") {
		t.Error("Set is missing value 'foo'")
	}
	if !s.ContainsString("bar") {
		t.Error("Set is missing value 'bar'")
	}
	if s.ContainsString("baz") {
		t.Error("Set is not missing value 'baz'")
	}
}

func TestNewLower(t *testing.T) {
	s := NewLowerFromString("Foo", "Bar", "foo")
	if !s.ContainsString("foo") {
		t.Error("Set is missing value 'foo'")
	}
	if !s.ContainsString("bar") {
		t.Error("Set is missing value 'bar'")
	}
	if s.ContainsString("baz") {
		t.Error("Set is not missing value 'baz'")
	}
}

func TestKeysLower(t *testing.T) {
	s := NewLowerFromString("Foo", "Bar", "baz")
	keys := s.Keys()
	if len(keys) != 3 {
		t.Error("Set keys list has incorrect size", keys)
	}
}

func TestPrintLower(t *testing.T) {
	x := NewLower()
	y := NewLowerFromString("a", "b")
	res1 := x.Print(",")
	res2 := y.Print(",")
	if res1 != "" {
		t.Error("Set print incorrect", res1)
	}
	if res2 != "a,b" {
		t.Error("Set print incorrect", res2)
	}
}

func TestIntersectionLower(t *testing.T) {
	a := NewLowerFromString("a", "B", "c")
	b := NewLowerFromString("b", "C", "y", "z")
	obs := a.Intersection(b)
	exp := NewLowerFromString("b", "c")
	if !reflect.DeepEqual(exp, obs) {
		t.Error("Set difference incorrect", obs)
	}
	obs = b.Intersection(a)
	if !reflect.DeepEqual(exp, obs) {
		t.Error("Set difference incorrect", obs)
	}
}

func TestJSONLower(t *testing.T) {
	var out Set
	in := NewLowerFromString("fOO", "Bar")
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
