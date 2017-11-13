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
package miniglob

import (
	"encoding/json"
	"testing"
)

func testPattern(t *testing.T, in string, out string) {
	obs := Pattern(in)
	if obs != out {
		t.Errorf("Pattern compilation of %s expected %s and observed %s",
			in, out, obs)
	}
}

func TestPattern(t *testing.T) {
	testPattern(t, "", "^$")
	testPattern(t, "foo", "^foo$")
	testPattern(t, "/foo/bar/", "^foo/bar$")
	testPattern(t, "*.java", "^[^/]*\\.java$")
	testPattern(t, "**.java", "^.*\\.java$")
	testPattern(t, "**/*.java", "^.*/[^/]*\\.java$")
}

func TestSerde(t *testing.T) {
	var body MiniGlob
	err := json.Unmarshal([]byte("\"  /hello*world/  \""), &body)
	if err != nil {
		t.Fatal("Unmarshal failure", err)
	}
	if body.Text != "  /hello*world/  " {
		t.Errorf("Unmarshal text context not successful: %s", body.Text)
	}
	if body.Regex.String() != "^hello[^/]*world$" {
		t.Errorf("Unmarshal regex context not successful: %s", body.Regex.String())
	}
	out, err := json.Marshal(body)
	if err != nil {
		t.Fatal("Marshal failure", err)
	}
	if string(out) != "\"  /hello*world/  \"" {
		t.Errorf("Marshal context not successful: %s", string(out))
	}
}
