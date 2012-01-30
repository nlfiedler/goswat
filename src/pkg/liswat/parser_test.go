//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package liswat

import (
	"testing"
)

func parserVerify(input, expected string, t *testing.T) {
	result, err := parseExpr(input)
	if err != nil {
		t.Errorf("failed to parse expression: " + input)
	}
	actual := stringify(result)
	if actual != expected {
		t.Errorf("expected '%s', but got '%s'", expected, actual)
	}
}

func TestParseExprEmptyList(t *testing.T) {
	input := "()"
	expected := "()"
	parserVerify(input, expected, t)
}

func TestParseExprSingletonList(t *testing.T) {
	input := "(foo)"
	expected := "(foo)"
	parserVerify(input, expected, t)
}

func TestParseExprList(t *testing.T) {
	input := "(foo  bar    baz)"
	expected := "(foo bar baz)"
	parserVerify(input, expected, t)
}

func TestParseExprNestedList(t *testing.T) {
	input := `(foo
  (bar
    baz))`
	expected := "(foo (bar baz))"
	parserVerify(input, expected, t)
}

func TestParseExprBoolean(t *testing.T) {
	input := "(#t #f)"
	expected := "(#t #f)"
	parserVerify(input, expected, t)
}

func TestParseExprString(t *testing.T) {
	input := `"foo"`
	expected := `"foo"`
	parserVerify(input, expected, t)
}

func TestParseExprFloat(t *testing.T) {
	input := "1.2345"
	expected := "1.2345"
	parserVerify(input, expected, t)
}

func TestParseExprInteger(t *testing.T) {
	input := "12345"
	expected := "12345"
	parserVerify(input, expected, t)
}
