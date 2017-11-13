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
package matcher

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildParseTree(t *testing.T) {
	input := "a and b or (us and them) or anyone and not d or f[self=true,count=10] and nof(a,b,c[self=false,count=2] and (d or e),1)"
	tokens := BuildTokens(input)
	_, err := BuildParseTree(tokens)
	assert.Nil(t, err)
}

func TestBuildParseTreeSmall(t *testing.T) {
	input := "a and b or c"
	tokens := BuildTokens(input)
	_, err := BuildParseTree(tokens)
	assert.Nil(t, err)
}

func TestBuildParseTreeParens(t *testing.T) {
	input := "(a and b) or not (c and (d or e))"
	tokens := BuildTokens(input)
	_, err := BuildParseTree(tokens)
	assert.Nil(t, err)
}

func TestBuildParseTreeInvalid(t *testing.T) {
	vals := []string{
		"(a and b) or not (c and (d or e)))",
		")",
		"(a and )",
		"( and a)",
	}
	for _, input := range vals {
		tokens := BuildTokens(input)
		_, err := BuildParseTree(tokens)
		assert.NotNil(t, err)
	}
}

func TestInvalidErrorMessages(t *testing.T) {
	tokens := BuildTokens("foo[and]")
	_, err := BuildParseTree(tokens)
	assert.Equal(t, err.Error(), "invalid 'and' at position 5")
	tokens = BuildTokens("foo[a=1")
	_, err = BuildParseTree(tokens)
	assert.Equal(t, err.Error(), "missing ']' at position 7")
	tokens = BuildTokens("foo[a,b,c]")
	_, err = BuildParseTree(tokens)
	assert.Equal(t, err.Error(), "invalid ',' at position 6")
}
