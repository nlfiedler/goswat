//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package liswat

import (
	"testing"
)

// expectedResult is equivalent to a token and is used in comparing the
// results from the lexer.
type expectedResult struct {
	typ tokenType
	val string
}

// verifyLexerResults calls lex() and checks that the resulting tokens
// match the expected results.
func verifyLexerResults(t *testing.T, input string, expected []expectedResult) {
	c := lex("unit", input)
	for i, e := range expected {
		token, ok := <-c
		if !ok {
			t.Errorf("lexer channel closed?")
		}
		if token.typ != e.typ {
			t.Errorf("expected %d, got %d for '%s' (token %d)", e.typ, token.typ, e.val, i)
		}
		if token.val != e.val {
			t.Errorf("expected '%s', got '%s' (token %d, type %d)", e.val, token.val, i, e.typ)
		}
	}
}

func TestLexerComments(t *testing.T) {
	input := `; foo
; bar baz
; quux
`
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenEOF, ""})
	verifyLexerResults(t, input, expected)
}

func TestLexerSimple(t *testing.T) {
	input := "(set foo bar)"
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenParen, "("})
	expected = append(expected, expectedResult{tokenVariable, "set"})
	expected = append(expected, expectedResult{tokenVariable, "foo"})
	expected = append(expected, expectedResult{tokenVariable, "bar"})
	expected = append(expected, expectedResult{tokenParen, ")"})
	expected = append(expected, expectedResult{tokenEOF, ""})
	verifyLexerResults(t, input, expected)
}

func TestLexerFactorial(t *testing.T) {
	input := `(define fact
    (lambda (n)
     (if (<= n 1)
         1
         (* n (fact (- n 1))))))
`
	expected := make([]expectedResult, 0)
	// 0
	expected = append(expected, expectedResult{tokenParen, "("})
	expected = append(expected, expectedResult{tokenVariable, "define"})
	expected = append(expected, expectedResult{tokenVariable, "fact"})
	expected = append(expected, expectedResult{tokenParen, "("})
	expected = append(expected, expectedResult{tokenVariable, "lambda"})
	// 5
	expected = append(expected, expectedResult{tokenParen, "("})
	expected = append(expected, expectedResult{tokenVariable, "n"})
	expected = append(expected, expectedResult{tokenParen, ")"})
	expected = append(expected, expectedResult{tokenParen, "("})
	expected = append(expected, expectedResult{tokenVariable, "if"})
	// 10
	expected = append(expected, expectedResult{tokenParen, "("})
	expected = append(expected, expectedResult{tokenVariable, "<="})
	expected = append(expected, expectedResult{tokenVariable, "n"})
	expected = append(expected, expectedResult{tokenInteger, "1"})
	expected = append(expected, expectedResult{tokenParen, ")"})
	// 15
	expected = append(expected, expectedResult{tokenInteger, "1"})
	expected = append(expected, expectedResult{tokenParen, "("})
	expected = append(expected, expectedResult{tokenVariable, "*"})
	expected = append(expected, expectedResult{tokenVariable, "n"})
	expected = append(expected, expectedResult{tokenParen, "("})
	// 20
	expected = append(expected, expectedResult{tokenVariable, "fact"})
	expected = append(expected, expectedResult{tokenParen, "("})
	expected = append(expected, expectedResult{tokenVariable, "-"})
	expected = append(expected, expectedResult{tokenVariable, "n"})
	expected = append(expected, expectedResult{tokenInteger, "1"})
	// 25
	expected = append(expected, expectedResult{tokenParen, ")"})
	expected = append(expected, expectedResult{tokenParen, ")"})
	expected = append(expected, expectedResult{tokenParen, ")"})
	expected = append(expected, expectedResult{tokenParen, ")"})
	expected = append(expected, expectedResult{tokenParen, ")"})
	// 30
	expected = append(expected, expectedResult{tokenParen, ")"})
	expected = append(expected, expectedResult{tokenEOF, ""})
	verifyLexerResults(t, input, expected)
}

func TestLexerUnclosedQuotes(t *testing.T) {
	input := `"foo`
	c := lex("unit", input)
	token, ok := <-c
	if !ok {
		t.Errorf("lexer channel closed?")
	}
	if token.typ != tokenError {
		t.Errorf("expected lexing unclosed quote to fail")
	}
}
