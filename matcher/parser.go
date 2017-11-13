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
	"fmt"

	"github.com/pkg/errors"
)

type parseKind int

const (
	_ parseKind = iota
	PARSE_NOUN
	PARSE_ANONYMOUS
	PARSE_JOINER
	PARSE_FUNCTION
)

type ParseToken interface {
	fmt.Stringer
	kind() parseKind
	parent() ParseToken
	setParent(ParseToken)
	isPopulated() bool
}

type JoinerKind int

const (
	_ JoinerKind = iota
	JOINER_AND
	JOINER_OR
	JOINER_NOT
)

type childToken struct {
	p ParseToken
}

func (ct *childToken) parent() ParseToken {
	return ct.p
}

func (ct *childToken) setParent(p ParseToken) {
	ct.p = p
}

type NounParseToken struct {
	Name       string
	Attributes map[string]string
	*childToken
}

func (*NounParseToken) kind() parseKind {
	return PARSE_NOUN
}

func (np *NounParseToken) String() string {
	label := np.Name
	return fmt.Sprintf("%s %v", label, np.Attributes)
}

func (*NounParseToken) isPopulated() bool {
	return true
}

type AnonymousParseToken struct {
	Members    []string
	Attributes map[string]string
	*childToken
}

func (*AnonymousParseToken) kind() parseKind {
	return PARSE_ANONYMOUS
}

func (np *AnonymousParseToken) String() string {
	return fmt.Sprintf("%v %v", np.Members, np.Attributes)
}

func (*AnonymousParseToken) isPopulated() bool {
	return true
}

type FunctionParseToken struct {
	Name       string
	Parameters []ParseToken
	*childToken
}

func (*FunctionParseToken) isPopulated() bool {
	return true
}

func (*FunctionParseToken) kind() parseKind {
	return PARSE_FUNCTION
}

func (fp *FunctionParseToken) String() string {
	label := fp.Name
	return fmt.Sprintf("%s %v", label, fp.Parameters)
}

type AndOrParseToken struct {
	JKind JoinerKind
	Left  ParseToken
	Right ParseToken
	*childToken
}

func (*AndOrParseToken) kind() parseKind {
	return PARSE_JOINER
}

func (aop *AndOrParseToken) String() string {
	l := "{INVALID}"
	if aop.Left != nil {
		l = aop.Left.String()
	}
	r := "{INVALID}"
	if aop.Right != nil {
		r = aop.Right.String()
	}
	j := "AND"
	if aop.JKind == JOINER_OR {
		j = "OR"
	}
	return fmt.Sprintf("%s %s %s", l, j, r)
}

func (aop *AndOrParseToken) isPopulated() bool {
	return aop.Left != nil && aop.Right != nil
}

type NotParseToken struct {
	Child ParseToken
	*childToken
}

func (*NotParseToken) kind() parseKind {
	return PARSE_JOINER
}

func (np *NotParseToken) String() string {
	if np.Child == nil {
		return "NOT {MISSING}"
	}
	return fmt.Sprintf("NOT %s", np.Child.String())
}

func (np *NotParseToken) isPopulated() bool {
	return np.Child != nil
}

type funcState int

const (
	_ funcState = iota
	FUNC_OUT
	FUNC_IN
)

/*
Valid grammar:

NOUN := NAME | US | THEM | ANY
PO_NAME := NOUN ATTRIBUTE_CLAUSE?
ATTRIBUTE_CLAUSE := LBRACKET (NAME EQUAL NAME COMMA)* NAME EQUAL NAME RBRACKET
NOT_STMT := NOT? PO_NAME
CLAUSE := NOT_STMT ((AND NOT_STMT) | (OR NOT_STMT))*

*/
type parseTreeBuilder struct {
	root      ParseToken
	lastToken ParseToken
}

func BuildParseTree(tokens []token) (ParseToken, error) {
	ptb := &parseTreeBuilder{}
	pos, err := ptb.buildParseTreeInner(tokens, FUNC_OUT)
	if err == nil && pos != len(tokens) {
		return nil, errors.New("premature end of tokens")
	}
	return ptb.root, err
}

func (ptb *parseTreeBuilder) addJoiner(kind JoinerKind) {
	pt := &AndOrParseToken{JKind: kind, childToken: &childToken{}}
	pt.setParent(ptb.lastToken.parent())
	if ptb.root == ptb.lastToken {
		ptb.root = pt
	}
	switch lt := ptb.lastToken.parent().(type) {
	case *AndOrParseToken:
		lt.Right = pt
	case *NotParseToken:
		lt.Child = pt
	}
	ptb.lastToken.setParent(pt)
	pt.Left = ptb.lastToken
	ptb.lastToken = pt
}

func (ptb *parseTreeBuilder) addToTree(pt ParseToken) {
	if ptb.root == nil {
		ptb.root = pt
	} else {
		switch lt := ptb.lastToken.(type) {
		case *AndOrParseToken:
			lt.Right = pt
		case *NotParseToken:
			lt.Child = pt
		default:
			panic(fmt.Sprintf("can't add a child to token %+v", ptb.lastToken))
		}
	}
	pt.setParent(ptb.lastToken)
	ptb.lastToken = pt
}

func (ptb *parseTreeBuilder) addNoun(name string) {
	pt := &NounParseToken{Name: name, childToken: &childToken{}}
	ptb.addToTree(pt)
}

func (ptb *parseTreeBuilder) addAnonymous(members []string) {
	pt := &AnonymousParseToken{Members: members, Attributes: map[string]string{}, childToken: &childToken{}}
	ptb.addToTree(pt)
}

func (ptb *parseTreeBuilder) addFunc(name string) {
	pt := &FunctionParseToken{Name: name, Parameters: []ParseToken{}, childToken: &childToken{}}
	ptb.addToTree(pt)
}

type attribState int

const (
	_ attribState = iota
	NONE
	IN_ATTRIB
	IN_NAME
	IN_EQ
	IN_VAL
	IN_COMMA
	IN_SET
	IN_SET_VAL
	IN_SET_COMMA
	OUT_SET
)

func makeError(t token) error {
	return errors.Errorf("invalid '%s' at position %d", t.name, t.pos)
}

func (ptb *parseTreeBuilder) handleAttributes(tokens []token, start int) (map[string]string, int, error) {
	if len(tokens) == 0 {
		return nil, 0, errors.Errorf("missing ']' to close '[' found at position %d", start)
	}
	attribs := map[string]string{}
	var curAttribName string
	inAttrib := IN_ATTRIB
	for i := 0; i < len(tokens); i++ {
		v := tokens[i]
		switch v.value {
		case TOKEN_NAME:
			switch inAttrib {
			case IN_ATTRIB, IN_COMMA:
				curAttribName = v.name
				inAttrib = IN_NAME
			case IN_EQ:
				attribs[curAttribName] = v.name
				inAttrib = IN_VAL
			default:
				return nil, 0, makeError(v)
			}

		case TOKEN_RBRACKET:
			if inAttrib == IN_VAL {
				return attribs, i + 1, nil
			}
			return nil, 0, makeError(v)

		case TOKEN_EQUAL:
			if inAttrib == IN_NAME {
				inAttrib = IN_EQ
			} else {
				return nil, 0, makeError(v)
			}

		case TOKEN_COMMA:
			if inAttrib == IN_VAL {
				inAttrib = IN_COMMA
			} else {
				return nil, 0, makeError(v)
			}
		default:
			return nil, 0, makeError(v)
		}
	}
	return nil, 0, errors.Errorf("missing ']' at position %d",
		tokens[len(tokens)-1].pos)
}

func (ptb *parseTreeBuilder) handleSetMembership(tokens []token, start int) ([]string, int, error) {
	var members []string
	if len(tokens) == 0 {
		return nil, 0, errors.Errorf("missing '}' to close '{' found at position %d", start)
	} else if tokens[0].value == TOKEN_RBRACE {
		return members, 1, nil
	}
	for i := 0; i < len(tokens)-1; i++ {
		t := tokens[i]
		if t.value != TOKEN_NAME {
			return nil, 0, makeError(t)
		}
		members = append(members, t.name)
		next := tokens[i+1].value
		if next == TOKEN_RBRACE {
			return members, i + 2, nil
		} else if next == TOKEN_COMMA {
			i++
		} else {
			return nil, 0, errors.Errorf("missing ',' at position %d", tokens[i+1].pos)
		}
	}
	return nil, 0, errors.Errorf("missing '}' at position %d",
		tokens[len(tokens)-1].pos)
}

func (ptb *parseTreeBuilder) handleFunctionArguments(tokens []token) ([]ParseToken, int, error) {
	done := false
	subPos := 0
	params := []ParseToken{}
	for !done {
		ptbInner := &parseTreeBuilder{}
		newPos, err := ptbInner.buildParseTreeInner(tokens[subPos:], FUNC_IN)
		if err != nil {
			return nil, 0, err
		}
		params = append(params, ptbInner.root)
		subPos += newPos
		switch tokens[subPos].value {
		case TOKEN_COMMA:
			subPos++
			continue //loop again
		case TOKEN_RPAREN:
			done = true //at end of parameters
		default:
			return nil, 0, makeError(tokens[subPos])
		}
	}

	//subPos should be pointing to an RParen, so skip over it and continue
	return params, subPos + 1, nil
}

func (ptb *parseTreeBuilder) beginAttributes() bool {
	last := ptb.lastToken
	if last == nil {
		return false
	}
	if last.kind() == PARSE_NOUN && len(last.(*NounParseToken).Attributes) == 0 {
		return true
	}
	if last.kind() == PARSE_ANONYMOUS && len(last.(*AnonymousParseToken).Attributes) == 0 {
		return true
	}
	return false
}

func (ptb *parseTreeBuilder) buildParseTreeInner(tokens []token, inFunc funcState) (int, error) {
	for pos := 0; pos < len(tokens); pos++ {
		v := tokens[pos]
		switch v.value {
		case TOKEN_LBRACE:
			members, newPos, err := ptb.handleSetMembership(tokens[pos+1:], v.pos)
			if err != nil {
				return 0, err
			}
			ptb.addAnonymous(members)
			pos += newPos
		case TOKEN_NAME:
			if ptb.lastToken != nil && ptb.lastToken.kind() != PARSE_JOINER {
				return 0, errors.Errorf("invalid noun %s at position %d", v.name, v.pos)
			}
			//figure out if this is a noun or a function by peeking at the next token
			if pos < len(tokens)-1 && tokens[pos+1].value == TOKEN_LPAREN {
				ptb.addFunc(v.name)
			} else {
				ptb.addNoun(v.name)
			}
		case TOKEN_LBRACKET:
			if !ptb.beginAttributes() {
				return 0, makeError(v)
			}
			attribs, newPos, err := ptb.handleAttributes(tokens[pos+1:], v.pos)
			if err != nil {
				return 0, err
			}
			switch lt := ptb.lastToken.(type) {
			case *NounParseToken:
				lt.Attributes = attribs
			case *AnonymousParseToken:
				lt.Attributes = attribs
			default:
				//this will trigger a panic but I'll know the type
				ptb.lastToken.(*NounParseToken).Attributes = attribs
			}
			pos += newPos
		case TOKEN_COMMA:
			if inFunc == FUNC_IN {
				return pos, nil
			}
			return 0, makeError(v)
		case TOKEN_AND:
			if ptb.lastToken == nil {
				return 0, makeError(v)
			}
			ptb.addJoiner(JOINER_AND)
		case TOKEN_OR:
			if ptb.lastToken == nil {
				return 0, makeError(v)
			}
			ptb.addJoiner(JOINER_OR)
		case TOKEN_NOT:
			if ptb.lastToken != nil && ptb.lastToken.kind() != PARSE_JOINER {
				return 0, makeError(v)
			}
			pt := &NotParseToken{childToken: &childToken{}}
			ptb.addToTree(pt)
		case TOKEN_LPAREN:
			if ptb.lastToken == nil || ptb.lastToken.kind() == PARSE_JOINER {
				//grouping parens, recurse (kinda)
				ptbInner := &parseTreeBuilder{}
				newPos, err := ptbInner.buildParseTreeInner(tokens[pos+1:], FUNC_OUT)
				if err != nil {
					return 0, err
				}
				ptb.addToTree(ptbInner.root)
				pos += newPos + 1
			} else if ptb.lastToken != nil && ptb.lastToken.kind() == PARSE_FUNCTION {
				//function parameters
				params, newPos, err := ptb.handleFunctionArguments(tokens[pos+1:])
				if err != nil {
					return 0, err
				}
				ptb.lastToken.(*FunctionParseToken).Parameters = params
				pos += newPos
			} else {
				return 0, makeError(v)
			}
		case TOKEN_RPAREN:
			if ptb.lastToken == nil || !ptb.lastToken.isPopulated() {
				return 0, makeError(v)
			}
			return pos, nil
		default:
			return 0, makeError(v)
		}
	}
	return len(tokens), nil
}
