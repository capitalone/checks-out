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
	"fmt"
	"regexp"
	"testing"

	"github.com/capitalone/checks-out/hjson"

	"github.com/stretchr/testify/assert"
)

func TestParseEmptyConfig(t *testing.T) {
	input := []byte{}
	_, err := ParseConfig(input, AllowAll())
	if err == nil {
		t.Fatal("Empty config did not generate error", err)
	}
}

func TestParseDefaultRegex(t *testing.T) {
	c := DefaultConfig()
	matcher := c.Pattern.Regex
	assert.Equal(t, `(?i)^I approve\s*(version:\s*(?P<version>\S+))?\s*(comment:\s*(?P<comment>.*\S))?\s*`, matcher.String())
	names := matcher.SubexpNames()
	fmt.Println(names)
	assert.Equal(t, 5, len(names))
	assert.Equal(t, "version", names[2])
	assert.Equal(t, "comment", names[4])
	assert.Nil(t, matcher.FindStringSubmatch("AYUP"))
	doRegEx(matcher, t, "I approve", []string{"I approve", "", ""})
	doRegEx(matcher, t, "I approve 1.2", []string{"I approve ", "", ""})
	doRegEx(matcher, t, "I approve version:", []string{"I approve ", "", ""})
	doRegEx(matcher, t, "I approve comment:", []string{"I approve ", "", ""})
	doRegEx(matcher, t, "I approve 1.2 hi there", []string{"I approve ", "", ""})
	doRegEx(matcher, t, "I approve version:1.2", []string{"I approve version:1.2", "1.2", ""})
	doRegEx(matcher, t, "I approve version: 1.2 comment: hi there", []string{"I approve version: 1.2 comment: hi there", "1.2", "hi there"})
	doRegEx(matcher, t, "I approve comment: hi there", []string{"I approve comment: hi there", "", "hi there"})
	doRegEx(matcher, t, "I approve comment:hi there ", []string{"I approve comment:hi there ", "", "hi there"})
}

func doRegEx(matcher *regexp.Regexp, t *testing.T, input string, values []string) {
	matches := matcher.FindStringSubmatch(input)
	fmt.Println(matches)
	assert.Equal(t, 5, len(matches))
	assert.Equal(t, values[0], matches[0])
	assert.Equal(t, values[1], matches[2])
	assert.Equal(t, values[2], matches[4])
}

func TestParseSimpleConfig(t *testing.T) {
	input := []byte(`
	approvals:
	[
	  {
	    match: "all[count=1,self=false]"
	  }
	]
	merge:
	{
	  enable: true
	}
	tag:
	{
	  enable: true
	}
	`)
	config, err := ParseConfig(input, AllowAll())
	if err != nil {
		t.Fatal("Unable to parse empty configuration", err)
	}
	assert.Equal(t, true, config.Tag.Enable)
	assert.Equal(t, "semver", config.Tag.Alg)
	assert.Equal(t, 1, len(config.Approvals))
	assert.Equal(t, 1, config.Approvals[0].Position)
	s, err := config.Tag.GenerateTag(TemplateTag{Version: "0.5.0"})
	assert.Nil(t, err)
	assert.Equal(t, "0.5.0", s)
}

func TestParseComplicatedConfig(t *testing.T) {
	input := []byte(`
approvals:
[
  {
    name: "dev"
    scope:
    {
      branches: [ "dev" ]
    }
    match: "all[count=1,self=true]"
    tag:
    {
      enable: true
      template: "{{.Version}}-dev"
      increment: "patch"
    }
  }
  {
    name: "master"
    scope:
    {
      branches: [ "master" ]
    }
    pattern: "(?i)^shipit\\s*(?P<version>\\S*)"
    match: "all[count=1,self=false]"
    tag:
    {
      enable: true
      template: "{{.Version}}"
      increment: "none"
    }
  }
  # Disable checks-out on remaining branches
  {
    match: "off"
  }
]
antipattern: "(?i)^holdit"
merge:
{
  enable: true
}
comment:
{
  enable: true
  targets: [
    {
      target: github
      types: [ "open" ]
    }
    {
      target: slack
      names: [ "#feed" ]
    }
  ]
}
	`)
	_, err := ParseConfig(input, AllowAll())
	if err != nil {
		t.Fatal("Unable to parse empty configuration", err)
	}
}

func TestOldConfig(t *testing.T) {
	input := []byte(`
	approvals = 3
	pattern = "LOOKGOOD"
	self_approval_off = true
	`)
	config, err := ParseOldConfig(input)
	if err != nil {
		t.Fatal("Expected valid file, got", err)
	}
	mm, ok := config.Approvals[0].Match.Matcher.(*MaintainerMatch)
	assert.True(t, ok, "Expected MaintainerMatch")
	assert.Equal(t, false, mm.Self)
	assert.Equal(t, 3, mm.Approvals)
	assert.Equal(t, "LOOKGOOD", config.Pattern.Regex.String())
}

func TestOldConfigDefault(t *testing.T) {
	input := []byte(`
	`)
	config, err := ParseOldConfig(input)
	if err != nil {
		t.Fatal("Expected valid file, got", err)
	}
	mm, ok := config.Approvals[0].Match.Matcher.(*MaintainerMatch)
	assert.True(t, ok, "Expected MaintainerMatch")
	assert.Equal(t, true, mm.Self)
	assert.Equal(t, 2, mm.Approvals)
	assert.Equal(t, "(?i)LGTM", config.Pattern.Regex.String())
}

func TestOldConfigConvert(t *testing.T) {
	input := []byte(`
	approvals = 3
	pattern = "LOOKGOOD"
	self_approval_off = true
	`)
	config, err := ParseOldConfig(input)
	if err != nil {
		t.Fatal("Expected valid file, got", err)
	}

	options := hjson.DefaultOptions()
	options.OmitEmptyStructs = true
	out, err := hjson.MarshalWithOptions(config, options)
	expected := `{
  approvals:
  [
    {
      scope:
      {
      }
      match: "all[count=3,self=false]"
      antimatch: "all[count=1,self=true]"
      authormatch: "universe[count=1,self=true]"
    }
  ]
  pattern: "LOOKGOOD"
  maintainers:
  {
    path: MAINTAINERS
    type: legacy
  }
  merge:
  {
    enable: false
    uptodate: true
    method: merge
    delete: false
  }
  feedback:
  {
    types:
    [
      "comment"
      "review"
    ]
    authoraffirm: true
  }
  audit:
  {
    enable: false
    branches: ["master"]
  }
}`
	if expected != string(out) {
		t.Fatalf("Expected\n%v\ngot\n%v\n", expected, string(out))
	}
}
