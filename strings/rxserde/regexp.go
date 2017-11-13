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
)

type RegexSerde struct {
	Regex *regexp.Regexp
}

func (rs RegexSerde) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	e := json.NewEncoder(&buf)
	e.SetEscapeHTML(false)
	err := e.Encode(rs.Regex.String())
	if err != nil {
		return nil, err
	}
	// golang json encoder adds a newline for each value
	return bytes.TrimSpace(buf.Bytes()), nil
}

func (rs *RegexSerde) UnmarshalJSON(text []byte) error {
	var pat string
	err := json.Unmarshal(text, &pat)
	if err != nil {
		return err
	}
	r, err := regexp.Compile(pat)
	if err != nil {
		return err
	}
	rs.Regex = r
	return nil
}
