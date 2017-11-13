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
package rxserde

import (
	"bytes"
	"encoding/json"
	"regexp"
	"testing"
)

type Testinner struct {
	Pattern RegexSerde
}

const text = `{
  "pattern": "abcdef"
}`

const expected = `"(?i)^I approve\\s*(?P<version>\\S*)"`

func TestMarshalDirect(t *testing.T) {
	r := RegexSerde{regexp.MustCompile(`(?i)^I approve\s*(?P<version>\S*)`)}
	b, err := r.MarshalJSON()
	if err != nil {
		t.Fatal("Unable to marshal regex serde", err)
	}
	s := string(b)
	if s != `"(?i)^I approve\\s*(?P<version>\\S*)"` {
		t.Error("marshal regex serde did not yield expected result", s)
	}
}

func TestMarshalRegexSerde(t *testing.T) {
	r := RegexSerde{regexp.MustCompile(`(?i)^I approve\s*(?P<version>\S*)`)}
	var buf bytes.Buffer
	e := json.NewEncoder(&buf)
	e.SetEscapeHTML(false)
	err := e.Encode(r)
	if err != nil {
		t.Fatal("Unable to marshal regex serde", err)
	}
	s := string(bytes.TrimSpace(buf.Bytes()))
	if s != `"(?i)^I approve\\s*(?P<version>\\S*)"` {
		t.Error("marshal regex serde did not yield expected result", s)
	}
}

func TestInnerRegexPattern(t *testing.T) {
	var result Testinner
	err := json.Unmarshal([]byte(text), &result)
	if err != nil {
		t.Fatal("Unable to unmarshal test struct", err)
	}
	if result.Pattern.Regex.String() != "abcdef" {
		t.Error("Unable to read inner pattern", result)
	}
}
