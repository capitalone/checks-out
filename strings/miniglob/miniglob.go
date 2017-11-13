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
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
)

type MiniGlob struct {
	Regex *regexp.Regexp
	Text  string
}

func Pattern(text string) string {
	var buffer bytes.Buffer
	buffer.WriteString("^")
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "/")
	text = strings.TrimSuffix(text, "/")
	outer := strings.Split(text, "**")
	for i, sec1 := range outer {
		inner := strings.Split(sec1, "*")
		for j, sec2 := range inner {
			buffer.WriteString(regexp.QuoteMeta(sec2))
			if j < len(inner)-1 {
				buffer.WriteString("[^/]*")
			}
		}
		if i < len(outer)-1 {
			buffer.WriteString(".*")
		}
	}
	buffer.WriteString("$")
	return buffer.String()
}

func Compile(text string) (*regexp.Regexp, error) {
	return regexp.Compile(Pattern(text))
}

func Create(text string) (MiniGlob, error) {
	regex, err := Compile(text)
	if err != nil {
		return MiniGlob{}, err
	}
	return MiniGlob{Text: text, Regex: regex}, nil
}

func MustCreate(text string) MiniGlob {
	glob, err := Create(text)
	if err != nil {
		panic(err)
	}
	return glob
}

func (rs MiniGlob) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	e := json.NewEncoder(&buf)
	e.SetEscapeHTML(false)
	err := e.Encode(rs.Text)
	if err != nil {
		return nil, err
	}
	// golang json encoder adds a newline for each value
	return bytes.TrimSpace(buf.Bytes()), nil
}

func (rs *MiniGlob) UnmarshalJSON(text []byte) error {
	var pat string
	err := json.Unmarshal(text, &pat)
	if err != nil {
		return err
	}
	r, err := Create(pat)
	if err != nil {
		return err
	}
	*rs = r
	return nil
}
