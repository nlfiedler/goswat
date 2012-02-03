//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package liswat

import (
	"strings"
	"testing"
)

// expectedLexerResult is equivalent to a token and is used in comparing the
// results from the lexer.
type expectedLexerResult struct {
	typ tokenType
	val string
}

type expectedLexerError struct {
	err string // expected error message substring
	msg string // explanation of error condition
}

// drainLexerChannel reads from the given channel until it closes.
func drainLexerChannel(c chan token) {
	for {
		_, ok := <-c
		if !ok {
			break
		}
	}
}

// verifyLexerResults calls lex() and checks that the resulting tokens
// match the expected results.
func verifyLexerResults(t *testing.T, input string, expected []expectedLexerResult) {
	c := lex("unit", input)
	for i, e := range expected {
		tok, ok := <-c
		if !ok {
			t.Errorf("lexer channel closed?")
		}
		if tok.typ != e.typ {
			t.Errorf("expected %d, got %d for '%s' (token %d)", e.typ, tok.typ, e.val, i)
		}
		if tok.val != e.val {
			t.Errorf("expected '%s', got '%s' (token %d, type %d)", e.val, tok.val, i, e.typ)
		}
	}
	drainLexerChannel(c)
}

// verifyLexerErrors calls lex() and checks that the resulting tokens
// resulted in an error, and (optionally) verifies the error message.
func verifyLexerErrors(t *testing.T, input map[string]expectedLexerError) {
	for i, e := range input {
		c := lex("unit", i)
		tok, ok := <-c
		if !ok {
			t.Errorf("lexer channel closed?")
		}
		if tok.typ != tokenError {
			t.Errorf("expected '%s' to fail with '%s'", i, e.err)
		}
		if !strings.Contains(tok.val, e.err) {
			t.Errorf("expected '%s' but got '%s'(%d) for input '%s'", e.err, tok.val, tok.typ, i)
		}
		drainLexerChannel(c)
	}
}

func TestLexerComments(t *testing.T) {
	input := `; foo
; bar baz
; quux
`
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenEOF, ""})
	verifyLexerResults(t, input, expected)
}

func TestLexerSimple(t *testing.T) {
	input := "(set foo bar)"
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "set"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "foo"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "bar"})
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	expected = append(expected, expectedLexerResult{tokenEOF, ""})
	verifyLexerResults(t, input, expected)
}

func TestLexerFactorial(t *testing.T) {
	input := `(define fact
    (lambda (n)
     (if (<= n 1)
         1
         (* n (fact (- n 1))))))
`
	expected := make([]expectedLexerResult, 0)
	// 0
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "define"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "fact"})
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "lambda"})
	// 5
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "n"})
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "if"})
	// 10
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "<="})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "n"})
	expected = append(expected, expectedLexerResult{tokenInteger, "1"})
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	// 15
	expected = append(expected, expectedLexerResult{tokenInteger, "1"})
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "*"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "n"})
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	// 20
	expected = append(expected, expectedLexerResult{tokenIdentifier, "fact"})
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "-"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "n"})
	expected = append(expected, expectedLexerResult{tokenInteger, "1"})
	// 25
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	// 30
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	expected = append(expected, expectedLexerResult{tokenEOF, ""})
	verifyLexerResults(t, input, expected)
}

func TestLexerUnclosedQuotes(t *testing.T) {
	input := make(map[string]expectedLexerError)
	input[`"foo`] = expectedLexerError{"unclosed quoted string", "unclosed quote should fail"}
	verifyLexerErrors(t, input)
}

func TestLexerQuotes(t *testing.T) {
	input := "'(foo bar) `(backtick) ,(baz qux) ,@(commat)"
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenQuote, "'"})
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "foo"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "bar"})
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	expected = append(expected, expectedLexerResult{tokenQuote, "`"})
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "backtick"})
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	expected = append(expected, expectedLexerResult{tokenQuote, ","})
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "baz"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "qux"})
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	expected = append(expected, expectedLexerResult{tokenQuote, ",@"})
	expected = append(expected, expectedLexerResult{tokenOpenParen, "("})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "commat"})
	expected = append(expected, expectedLexerResult{tokenCloseParen, ")"})
	verifyLexerResults(t, input, expected)
}

func TestLexerCharacters(t *testing.T) {
	input := "#\\a #\\space #\\newline #\\t"
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenCharacter, "#\\a"})
	expected = append(expected, expectedLexerResult{tokenCharacter, "#\\ "})
	expected = append(expected, expectedLexerResult{tokenCharacter, "#\\\n"})
	expected = append(expected, expectedLexerResult{tokenCharacter, "#\\t"})
	verifyLexerResults(t, input, expected)
}

func TestLexerBadCharacter(t *testing.T) {
	input := make(map[string]expectedLexerError)
	input["#\\abc"] = expectedLexerError{"malformed character escape", "invalid char escape should fail"}
	input["#\\0"] = expectedLexerError{"malformed character escape", "invalid char escape should fail"}
	input["#\\a1"] = expectedLexerError{"malformed character escape", "invalid char escape should fail"}
	verifyLexerErrors(t, input)
}

func TestLexerNumbers(t *testing.T) {
	input := ".01 0 0.1 1.00 123 6e4 7.91e+16 0366 0x7b5 3."
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenFloat, ".01"})
	expected = append(expected, expectedLexerResult{tokenInteger, "0"})
	expected = append(expected, expectedLexerResult{tokenFloat, "0.1"})
	expected = append(expected, expectedLexerResult{tokenFloat, "1.00"})
	expected = append(expected, expectedLexerResult{tokenInteger, "123"})
	expected = append(expected, expectedLexerResult{tokenFloat, "6e4"})
	expected = append(expected, expectedLexerResult{tokenFloat, "7.91e+16"})
	expected = append(expected, expectedLexerResult{tokenInteger, "0366"})
	expected = append(expected, expectedLexerResult{tokenInteger, "0x7b5"})
	expected = append(expected, expectedLexerResult{tokenFloat, "3."})
	verifyLexerResults(t, input, expected)
}

func TestLexerBadNumbers(t *testing.T) {
	input := make(map[string]expectedLexerError)
	input["0.a"] = expectedLexerError{"malformed number", "invalid number should fail"}
	input["0a"] = expectedLexerError{"malformed number", "invalid number should fail"}
	verifyLexerErrors(t, input)
}

func TestLexerIdentifiers(t *testing.T) {
	input := "lambda list->vector q soup V17a + <=? a34kTMNs the-word-recursion-has-many-meanings - . ... "
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenIdentifier, "lambda"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "list->vector"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "q"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "soup"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "V17a"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "+"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "<=?"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "a34kTMNs"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "the-word-recursion-has-many-meanings"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "-"})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "."})
	expected = append(expected, expectedLexerResult{tokenIdentifier, "..."})
	verifyLexerResults(t, input, expected)
}

func TestLexerBadIdentifiers(t *testing.T) {
	input := make(map[string]expectedLexerError)
	input[".a"] = expectedLexerError{"malformed identifier", "invalid identifier should fail"}
	input["+a"] = expectedLexerError{"malformed identifier", "invalid identifier should fail"}
	input["-a"] = expectedLexerError{"malformed identifier", "invalid identifier should fail"}
	input[".. "] = expectedLexerError{"malformed identifier", "invalid identifier should fail"}
	input["...a"] = expectedLexerError{"malformed identifier", "invalid identifier should fail"}
	input[".... "] = expectedLexerError{"malformed identifier", "invalid identifier should fail"}
	verifyLexerErrors(t, input)
}
