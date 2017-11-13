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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasics(t *testing.T) {
	tokens := BuildTokens("a and b")
	root, err := BuildParseTree(tokens)
	assert.Nil(t, err)
	//should have a tree with a root of andornode and a left of noun node a with no attribs and a right of noun node b with no attribs
	andNode, ok := root.(*AndOrParseToken)
	assert.True(t, ok)
	assert.Equal(t, andNode.JKind, JOINER_AND)
	aNode, ok := andNode.Left.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, aNode.Name, "a")
	assert.Equal(t, 0, len(aNode.Attributes))

	bNode, ok := andNode.Right.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, bNode.Name, "b")
	assert.Equal(t, 0, len(bNode.Attributes))
}

func TestAnonymousGroup(t *testing.T) {
	tokens := BuildTokens("{}")
	root, err := BuildParseTree(tokens)
	assert.Nil(t, err)
	node, ok := root.(*AnonymousParseToken)
	assert.True(t, ok)
	assert.Equal(t, 0, len(node.Members))
	tokens = BuildTokens("{a, b,c }")
	root, err = BuildParseTree(tokens)
	assert.Nil(t, err)
	node, ok = root.(*AnonymousParseToken)
	assert.True(t, ok)
	assert.Equal(t, 3, len(node.Members))
}

func TestAttributes(t *testing.T) {
	tokens := BuildTokens("a and b[count=1,self=false]")
	root, err := BuildParseTree(tokens)
	assert.Nil(t, err)
	checkTestAttributes(root, t)

	tokens = BuildTokens("a and b[self=false,count=1]")
	root, err = BuildParseTree(tokens)
	assert.Nil(t, err)
	checkTestAttributes(root, t)
}

func checkTestAttributes(root ParseToken, t *testing.T) {
	//should have a tree with a root of andornode and a left of noun node a with no attribs and a right of noun node b with expected attribs
	andNode, ok := root.(*AndOrParseToken)
	assert.True(t, ok)
	assert.Equal(t, andNode.JKind, JOINER_AND)
	aNode, ok := andNode.Left.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, aNode.Name, "a")
	assert.Equal(t, 0, len(aNode.Attributes))

	bNode, ok := andNode.Right.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, bNode.Name, "b")
	assert.Equal(t, 2, len(bNode.Attributes))
	assert.Equal(t, "1", bNode.Attributes["count"])
	assert.Equal(t, "false", bNode.Attributes["self"])
}

func TestNestedSimple(t *testing.T) {
	tokens := BuildTokens("(a and b)")
	root, err := BuildParseTree(tokens)
	assert.Nil(t, err)
	checkNestedSimple(root, t)

	tokens = BuildTokens("((a) and b)")
	root, err = BuildParseTree(tokens)
	assert.Nil(t, err)
	checkNestedSimple(root, t)

	tokens = BuildTokens("(a and (b))")
	root, err = BuildParseTree(tokens)
	assert.Nil(t, err)
	checkNestedSimple(root, t)
}

func checkNestedSimple(root ParseToken, t *testing.T) {
	//should have a tree with a root of andornode and a left of noun node a with no attribs and a right of noun node b with no attribs
	andNode, ok := root.(*AndOrParseToken)
	assert.True(t, ok)
	assert.Equal(t, andNode.JKind, JOINER_AND)
	aNode, ok := andNode.Left.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, aNode.Name, "a")
	assert.Equal(t, 0, len(aNode.Attributes))

	bNode, ok := andNode.Right.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, bNode.Name, "b")
	assert.Equal(t, 0, len(bNode.Attributes))
}

func TestNestedComplex(t *testing.T) {
	tokens := BuildTokens("a and b or c")
	root, err := BuildParseTree(tokens)
	assert.Nil(t, err)
	checkNestedComplex1(root, t)

	tokens = BuildTokens("(a and b) or c")
	root, err = BuildParseTree(tokens)
	assert.Nil(t, err)
	checkNestedComplex2(root, t)

	tokens = BuildTokens("a and (b or c)")
	root, err = BuildParseTree(tokens)
	assert.Nil(t, err)
	checkNestedComplex1(root, t)
}

func checkNestedComplex1(root ParseToken, t *testing.T) {
	/*
	     and
	    /   \
	   a    or
	        / \
	       b   c
	*/
	andNode, ok := root.(*AndOrParseToken)
	assert.True(t, ok)
	assert.Equal(t, andNode.JKind, JOINER_AND)

	aNode, ok := andNode.Left.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, aNode.Name, "a")
	assert.Equal(t, 0, len(aNode.Attributes))

	orNode, ok := andNode.Right.(*AndOrParseToken)
	assert.True(t, ok)
	assert.Equal(t, orNode.JKind, JOINER_OR)

	bNode, ok := orNode.Left.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, bNode.Name, "b")
	assert.Equal(t, 0, len(bNode.Attributes))

	cNode, ok := orNode.Right.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, cNode.Name, "c")
	assert.Equal(t, 0, len(cNode.Attributes))
}

func checkNestedComplex2(root ParseToken, t *testing.T) {
	/*
	       or
	      /   \
	    and    c
	    / \
	   a   b
	*/
	orNode, ok := root.(*AndOrParseToken)
	assert.True(t, ok)
	assert.Equal(t, orNode.JKind, JOINER_OR)

	andNode, ok := orNode.Left.(*AndOrParseToken)
	assert.True(t, ok)
	assert.Equal(t, andNode.JKind, JOINER_AND)

	aNode, ok := andNode.Left.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, aNode.Name, "a")
	assert.Equal(t, 0, len(aNode.Attributes))

	bNode, ok := andNode.Right.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, bNode.Name, "b")
	assert.Equal(t, 0, len(bNode.Attributes))

	cNode, ok := orNode.Right.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, cNode.Name, "c")
	assert.Equal(t, 0, len(cNode.Attributes))
}

func TestNot(t *testing.T) {
	tokens := BuildTokens("not a")
	root, err := BuildParseTree(tokens)
	assert.Nil(t, err)
	checkNot1(root, t)

	tokens = BuildTokens("not not a")
	root, err = BuildParseTree(tokens)
	assert.Nil(t, err)
	checkNot2(root, t)

	tokens = BuildTokens("b and not not a")
	root, err = BuildParseTree(tokens)
	assert.Nil(t, err)
	checkNot3(root, t)
}

func checkNot1(root ParseToken, t *testing.T) {
	notNode, ok := root.(*NotParseToken)
	assert.True(t, ok)

	aNode, ok := notNode.Child.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, aNode.Name, "a")
	assert.Equal(t, 0, len(aNode.Attributes))
}

func checkNot2(root ParseToken, t *testing.T) {
	notNode, ok := root.(*NotParseToken)
	assert.True(t, ok)

	notNode2, ok := notNode.Child.(*NotParseToken)
	assert.True(t, ok)

	aNode, ok := notNode2.Child.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, aNode.Name, "a")
	assert.Equal(t, 0, len(aNode.Attributes))
}

func checkNot3(root ParseToken, t *testing.T) {
	andNode, ok := root.(*AndOrParseToken)
	assert.True(t, ok)
	assert.Equal(t, andNode.JKind, JOINER_AND)

	bNode, ok := andNode.Left.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, bNode.Name, "b")
	assert.Equal(t, 0, len(bNode.Attributes))

	notNode, ok := andNode.Right.(*NotParseToken)
	assert.True(t, ok)

	notNode2, ok := notNode.Child.(*NotParseToken)
	assert.True(t, ok)

	aNode, ok := notNode2.Child.(*NounParseToken)
	assert.True(t, ok)
	assert.Equal(t, aNode.Name, "a")
	assert.Equal(t, 0, len(aNode.Attributes))
}

func TestFunctionsSimple(t *testing.T) {
	tokens := BuildTokens("a(x,y,z,1)")
	root, err := BuildParseTree(tokens)
	assert.Nil(t, err)
	checkFunctionsSimple1(root, t)
}

func checkFunctionsSimple1(root ParseToken, t *testing.T) {
	aNode, ok := root.(*FunctionParseToken)
	assert.True(t, ok)
	assert.Equal(t, aNode.Name, "a")
	params := aNode.Parameters
	assert.Equal(t, 4, len(params))
	if len(params) == 4 {
		xNode, ok := params[0].(*NounParseToken)
		assert.True(t, ok)
		assert.Equal(t, "x", xNode.Name)

		yNode, ok := params[1].(*NounParseToken)
		assert.True(t, ok)
		assert.Equal(t, "y", yNode.Name)

		zNode, ok := params[2].(*NounParseToken)
		assert.True(t, ok)
		assert.Equal(t, "z", zNode.Name)

		oneNode, ok := params[3].(*NounParseToken)
		assert.True(t, ok)
		assert.Equal(t, "1", oneNode.Name)
	}
}

func TestFunctionsNested(t *testing.T) {
	tokens := BuildTokens("a(x[self=false],y(3),(z and b(2,3)),1)")
	root, err := BuildParseTree(tokens)
	assert.Nil(t, err)
	checkFunctionsNested(root, t)
}

func checkFunctionsNested(root ParseToken, t *testing.T) {
	aNode, ok := root.(*FunctionParseToken)
	assert.True(t, ok)
	assert.Equal(t, aNode.Name, "a")
	params := aNode.Parameters
	assert.Equal(t, 4, len(params))

	if len(params) == 4 {
		xNode, ok := params[0].(*NounParseToken)
		assert.True(t, ok)
		assert.Equal(t, "x", xNode.Name)
		assert.Equal(t, 1, len(xNode.Attributes))
		assert.Equal(t, "false", xNode.Attributes["self"])

		yNode, ok := params[1].(*FunctionParseToken)
		assert.True(t, ok)
		assert.Equal(t, "y", yNode.Name)
		assert.Equal(t, 1, len(yNode.Parameters))
		assert.Equal(t, "3", yNode.Parameters[0].(*NounParseToken).Name)

		andNode, ok := params[2].(*AndOrParseToken)
		assert.True(t, ok)
		assert.Equal(t, andNode.JKind, JOINER_AND)

		zNode, ok := andNode.Left.(*NounParseToken)
		assert.True(t, ok)
		assert.Equal(t, "z", zNode.Name)

		bNode, ok := andNode.Right.(*FunctionParseToken)
		assert.True(t, ok)
		assert.Equal(t, "b", bNode.Name)
		assert.Equal(t, 2, len(bNode.Parameters))

		assert.Equal(t, "2", bNode.Parameters[0].(*NounParseToken).Name)
		assert.Equal(t, "3", bNode.Parameters[1].(*NounParseToken).Name)

		oneNode, ok := params[3].(*NounParseToken)
		assert.True(t, ok)
		assert.Equal(t, "1", oneNode.Name)
	}
}
