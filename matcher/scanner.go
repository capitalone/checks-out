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

/*
language parse tokens:

NAME - refers to: a person_name, an org_name, a special org name (us, them, any), an attribute name, an attribute value, a number, true, or false
LBRACKET - [ for specifying attributes
RBRACKET - ] for specifying attributes
LPAREN - ( for grouping terms
RPAREN - ) for grouping terms
EQUAL - = for assigning attribute value to an attribute name
COMMA - , for separating items in attribute or parameter lists
AND - logical and
OR - logical or
NOT - logical not
LBRACE - { for specifying anonymous groups
RBRACE - } for specifying anonymous groups
*/

type token struct {
	name  string
	value tokenType
	pos   int
}

type tokenType int

const (
	TOKEN_INVALID tokenType = iota
	TOKEN_NAME
	TOKEN_LBRACKET
	TOKEN_RBRACKET
	TOKEN_EQUAL
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_COMMA
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	TOKEN_LBRACE
	TOKEN_RBRACE
)

func getTokenType(curString string) tokenType {
	switch curString {
	case "and":
		return TOKEN_AND
	case "or":
		return TOKEN_OR
	case "not":
		return TOKEN_NOT
	default:
		return TOKEN_NAME
	}
}

func BuildTokens(in string) []token {
	var curString []rune
	var curPos int
	out := []token{}
	f := func(t token) {
		if len(curString) > 0 {
			cur := string(curString)
			out = append(out, token{cur, getTokenType(cur), curPos})
			curString = curString[:0]
			curPos = 0
		}
		if len(t.name) > 0 {
			out = append(out, t)
		}
	}
	for i, v := range in {
		pos := i + 1
		switch v {
		case '[':
			f(token{"[", TOKEN_LBRACKET, pos})
		case ']':
			f(token{"]", TOKEN_RBRACKET, pos})
		case '=':
			f(token{"=", TOKEN_EQUAL, pos})
		case '(':
			f(token{"(", TOKEN_LPAREN, pos})
		case ')':
			f(token{")", TOKEN_RPAREN, pos})
		case '{':
			f(token{"{", TOKEN_LBRACE, pos})
		case '}':
			f(token{"}", TOKEN_RBRACE, pos})
		case ',':
			f(token{",", TOKEN_COMMA, pos})
		case ' ', '\n', '\t', '\r':
			f(token{"", 0, 0})
		default:
			if curPos == 0 {
				curPos = pos
			}
			curString = append(curString, v)
		}
	}
	f(token{"", 0, 0})
	return out
}
