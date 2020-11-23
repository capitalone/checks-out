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
package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"text/template"
)

type Tag string

type TagConfig struct {
	// enable automatic tagging of merges. Default is false.
	Enable bool `json:"enable"`
	// algorithm for generating tag name. Allowed values are
	// "explicit", "semver", "timestamp-rfc3339", and "timestamp-millis".
	// Default is "semver".
	Alg string `json:"algorithm"`
	// golang text/template for producing tag name.
	// Container struct is TemplateTag.
	// Default template is "{{.Version}}"
	TemplateRaw string             `json:"template"`
	Template    *template.Template `json:"-"`
	// If true then perform stricter validation
	// of templates to comply with Docker tag requirements
	Docker bool `json:"docker"`
	// Version increment policy for the "semver" algorithm.
	// Allowed values are "major", "minor", "patch", and "none".
	// Default is "patch".
	Increment Semver `json:"increment"`
}

type TemplateTag struct {
	Version string
}

func DefaultTag() TagConfig {
	return TagConfig{
		Enable:      false,
		Alg:         "semver",
		TemplateRaw: "{{.Version}}",
		Docker:      false,
		Increment:   Patch,
	}
}

var dockerRegex = regexp.MustCompile(`^[A-Za-z0-9_.\-]+$`)

var illegalRegex = regexp.MustCompile(`[[:cntrl:]]|[ ~^:?*\\\]]`)

// Used to avoid recursion in UnmarshalJSON
type shadowTagConfig TagConfig

func (t *TagConfig) UnmarshalJSON(text []byte) error {
	dummy := shadowTagConfig(DefaultTag())
	err := json.Unmarshal(text, &dummy)
	if err != nil {
		return err
	}
	*t = TagConfig(dummy)
	return nil
}

func checkRefFormat(t *TagConfig, tag string) error {
	ills := illegalRegex.FindString(tag)
	if len(ills) > 0 {
		err := fmt.Errorf(`Illegal template tag %s:
			cannot have the illegal character %s. Illegal characters are
			ASCII control character, space, tilde, caret, colon,
			question-mark, asterisk, open bracket,
			and blackslash`, t.TemplateRaw, ills)
		return err
	}
	return nil
}

func (t *TagConfig) execute(body interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	err := t.Template.Execute(&buffer, body)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (t *TagConfig) Compile() error {
	err := t.build()
	if err != nil {
		return err
	}
	return t.validate()
}

func (t *TagConfig) build() error {
	var err error
	t.Template, err = template.New("tag").Parse(t.TemplateRaw)
	return err
}

func (t *TagConfig) validate() error {
	buffer, err := t.execute(TemplateTag{Version: "1.0.0"})
	if err != nil {
		return err
	}
	tpl := string(buffer)
	if t.Docker && !dockerRegex.MatchString(tpl) {
		err := fmt.Errorf(`Illegal template tag %s with Docker validation enabled:
			only [A-Za-z0-9_.-] characters are allowed`, t.TemplateRaw)
		return err
	}
	return checkRefFormat(t, tpl)
}

func (t *TagConfig) GenerateTag(body TemplateTag) (string, error) {
	buffer, err := t.execute(body)
	if err != nil {
		return "", err
	}
	return string(buffer), nil
}
