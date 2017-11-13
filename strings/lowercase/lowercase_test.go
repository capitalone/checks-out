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
package lowercase

import (
	"testing"
)

func TestCreate(t *testing.T) {
	l := Create("Foobar")
	if l.val != "foobar" {
		t.Error("Lowercase value generated incorrectly")
	}
}

func TestLowercase(t *testing.T) {
	l := Create("Foobar")
	if l.String() != "foobar" {
		t.Error("Lowercase value retrieved incorrectly")
	}
}

func TestEmpty(t *testing.T) {
	var l String
	if l.String() != "" {
		t.Error("empty lowercase not handled correctly")
	}
}