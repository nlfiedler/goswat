//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"testing"
)

// parserResult is the expected state of the parser.
type parserResult struct {
	state parserState
	token parserToken
	len int
	start int
	end int
	p int
	quote bool
}

// validateParser checks that the parser is in exactly the state
// specified by the result, failing if that is not the case.
func validateParser(state parserState, parser *Parser, result parserResult, t *testing.T) {
	if state != result.state {
		t.Errorf("state did not match (expected %d, actual %d)", result.state, state)
	}
	if parser.token != result.token {
		t.Errorf("token did not match (expected %d, actual %d)", result.token, parser.token)
	}
	if parser.start != result.start {
		t.Errorf("start did not match (expected %d, actual %d)", result.start, parser.start)
	}
	if parser.p != result.p {
		t.Errorf("p did not match (expected %d, actual %d)", result.p, parser.p)
	}
	if parser.end != result.end {
		t.Errorf("end did not match (expected %d, actual %d)", result.end, parser.end)
	}
	if parser.len != result.len {
		t.Errorf("len did not match (expected %d, actual %d)", result.len, parser.len)
	}
	if parser.insidequote != result.quote {
		t.Errorf("insidequote did not match (expected %t, actual %t)", result.quote, parser.insidequote)
	}
}

func TestParseNewParser(t *testing.T) {
	parser := NewParser("foobar")
	result := parserResult{stateOK, tokenEOL, 6, 0, 0, 0, false}
	validateParser(stateOK, parser, result, t)
}

//
// parseSep
//

func TestParseSepAllSpace(t *testing.T) {
	parser := NewParser("  \n  \r \t ")
	state, _ := parser.parseSep()
	result := parserResult{stateOK, tokenSeparator, 0, 0, 8, 9, false}
	validateParser(state, parser, result, t)
}

func TestParseSepNoSpace(t *testing.T) {
	// technically this is an invalid starting state...
	parser := NewParser("foobar")
	state, _ := parser.parseSep()
	// ...and hence the end is a strange value
	result := parserResult{stateOK, tokenSeparator, 6, 0, -1, 0, false}
	validateParser(state, parser, result, t)
}

func TestParseSepSomeSpace(t *testing.T) {
	parser := NewParser("   bar")
	state, _ := parser.parseSep()
	result := parserResult{stateOK, tokenSeparator, 3, 0, 2, 3, false}
	validateParser(state, parser, result, t)
}

//
// parseEol
//

func TestParseEolAllSpace(t *testing.T) {
	parser := NewParser(" ; \n ; \r \t ;")
	state, _ := parser.parseEol()
	result := parserResult{stateOK, tokenEOL, 0, 0, 11, 12, false}
	validateParser(state, parser, result, t)
}

func TestParseEolNoSpace(t *testing.T) {
	// technically this is an invalid starting state...
	parser := NewParser("foobar")
	state, _ := parser.parseEol()
	// ...and hence the end is a strange value
	result := parserResult{stateOK, tokenEOL, 6, 0, -1, 0, false}
	validateParser(state, parser, result, t)
}

func TestParseEolSomeSpace(t *testing.T) {
	parser := NewParser("bar ; ")
	parser.p = 3
	parser.len = 3
	state, _ := parser.parseEol()
	result := parserResult{stateOK, tokenEOL, 0, 3, 5, 6, false}
	validateParser(state, parser, result, t)
}

//
// parseComment
//

func TestParseCommentNoEol(t *testing.T) {
	parser := NewParser("# foobar")
	state, _ := parser.parseComment()
	result := parserResult{stateOK, tokenEOL, 0, 0, 0, 8, false}
	validateParser(state, parser, result, t)
}

func TestParseCommentNewline(t *testing.T) {
	parser := NewParser("# foobar\n")
	state, _ := parser.parseComment()
	result := parserResult{stateOK, tokenEOL, 1, 0, 0, 8, false}
	validateParser(state, parser, result, t)
}

func TestParseCommentReturn(t *testing.T) {
	parser := NewParser("# foobar\r")
	state, _ := parser.parseComment()
	result := parserResult{stateOK, tokenEOL, 1, 0, 0, 8, false}
	validateParser(state, parser, result, t)
}

//
// parseCommand
//

func TestParseCommand(t *testing.T) {
	parser := NewParser("[foo {bar baz} quux]")
	state, _ := parser.parseCommand()
	result := parserResult{stateOK, tokenCommand, 0, 1, 18, 20, false}
	validateParser(state, parser, result, t)
}

func TestParseCommandSuffix(t *testing.T) {
	parser := NewParser("[foo {bar baz} quux]; # foo")
	state, _ := parser.parseCommand()
	result := parserResult{stateOK, tokenCommand, 7, 1, 18, 20, false}
	validateParser(state, parser, result, t)
}

//
// parseVariable
//

func TestParseVariable(t *testing.T) {
	parser := NewParser("$foo")
	state, _ := parser.parseVariable()
	result := parserResult{stateOK, tokenVariable, 0, 1, 3, 4, false}
	validateParser(state, parser, result, t)
}

func TestParseVariableBraces(t *testing.T) {
	parser := NewParser("${foo}bar")
	state, _ := parser.parseVariable()
	result := parserResult{stateOK, tokenVariable, 3, 2, 4, 6, false}
	validateParser(state, parser, result, t)
}

func TestParseVariableUnmatchedBrace(t *testing.T) {
	parser := NewParser("${foobar")
	state, err := parser.parseVariable()
	if state != stateError {
		t.Error("expected error state")
	}
	if err.Errno != EBRACE {
		t.Error("expected unmatched brace error")
	}
}

func TestParseVariableSpace(t *testing.T) {
	parser := NewParser("$foo ")
	state, _ := parser.parseVariable()
	result := parserResult{stateOK, tokenVariable, 1, 1, 3, 4, false}
	validateParser(state, parser, result, t)
}

func TestParseVariableDollar(t *testing.T) {
	parser := NewParser("$ ")
	state, _ := parser.parseVariable()
	result := parserResult{stateOK, tokenString, 1, 0, 0, 1, false}
	validateParser(state, parser, result, t)
}

//
// parseBrace
//

func TestParseBrace(t *testing.T) {
	parser := NewParser("{foo}")
	state, _ := parser.parseBrace()
	result := parserResult{stateOK, tokenString, 0, 1, 3, 5, false}
	validateParser(state, parser, result, t)
}

func TestParseBraceEmpty(t *testing.T) {
	parser := NewParser("{}")
	state, _ := parser.parseBrace()
	result := parserResult{stateOK, tokenString, 0, 1, 0, 2, false}
	validateParser(state, parser, result, t)
}

func TestParseBraceNested(t *testing.T) {
	parser := NewParser("{foo{bar}}")
	state, _ := parser.parseBrace()
	result := parserResult{stateOK, tokenString, 0, 1, 8, 10, false}
	validateParser(state, parser, result, t)
}

func TestParseBraceEscaped(t *testing.T) {
	parser := NewParser("{foo\\{bar\\}}")
	state, _ := parser.parseBrace()
	result := parserResult{stateOK, tokenString, 0, 1, 10, 12, false}
	validateParser(state, parser, result, t)
}

//
// parseString
//

func TestParseStringEmptyBraces(t *testing.T) {
	parser := NewParser("{}")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenString, 0, 1, 0, 2, false}
	validateParser(state, parser, result, t)
}

func TestParseStringNestedBraces(t *testing.T) {
	parser := NewParser("{foo{bar}}")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenString, 0, 1, 8, 10, false}
	validateParser(state, parser, result, t)
}

func TestParseStringEscapedBraces(t *testing.T) {
	parser := NewParser("{foo\\{bar\\}}")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenString, 0, 1, 10, 12, false}
	validateParser(state, parser, result, t)
}

func TestParseString(t *testing.T) {
	parser := NewParser("\"foobar\"")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenEscape, 0, 1, 6, 8, false}
	validateParser(state, parser, result, t)
}

func TestParseStringEscapes(t *testing.T) {
	parser := NewParser("\"f\\to;o\\\"b\\na\\rr\"")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenEscape, 0, 1, 15, 17, false}
	validateParser(state, parser, result, t)
}

func TestParseStringSeparators(t *testing.T) {
	parser := NewParser("foo bar")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenEscape, 4, 0, 2, 3, false}
	validateParser(state, parser, result, t)
}

func TestParseStringVariable(t *testing.T) {
	parser := NewParser("\"foo $bar\"")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenEscape, 5, 1, 4, 5, true}
	validateParser(state, parser, result, t)
}

func TestParseStringCommand(t *testing.T) {
	parser := NewParser("\"foo [bar]\"")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenEscape, 6, 1, 4, 5, true}
	validateParser(state, parser, result, t)
}

//
// GetTokenText
//

func TestGetTokenWord(t *testing.T) {
	parser := NewParser("foobar")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenEscape, 0, 0, 5, 6, false}
	validateParser(state, parser, result, t)
	if parser.GetTokenText() != "foobar" {
		t.Error("GetTokenText failed")
	}
}

func TestGetTokenVariable(t *testing.T) {
	parser := NewParser("$foobar")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenVariable, 0, 1, 6, 7, false}
	validateParser(state, parser, result, t)
	if parser.GetTokenText() != "foobar" {
		t.Error("GetTokenText failed")
	}
}

func TestGetTokenVariableBraces(t *testing.T) {
	parser := NewParser("${foo}bar")
	state, _ := parser.parseVariable()
	result := parserResult{stateOK, tokenVariable, 3, 2, 4, 6, false}
	validateParser(state, parser, result, t)
	if parser.GetTokenText() != "foo" {
		t.Error("GetTokenText failed")
	}
}

//
// parseToken
//

func TestParseTokenBlank(t *testing.T) {
	parser := NewParser("")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenEOF, 0, 0, 0, 0, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenWord(t *testing.T) {
	parser := NewParser("foobar")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenEscape, 0, 0, 5, 6, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenVariable(t *testing.T) {
	parser := NewParser("$foobar")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenVariable, 0, 1, 6, 7, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenEol(t *testing.T) {
	parser := NewParser("; \n \t \r")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenEOL, 0, 0, 6, 7, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenComment(t *testing.T) {
	parser := NewParser("# foo")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenEOF, 0, 0, 0, 5, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenEmptyBraces(t *testing.T) {
	parser := NewParser("{}")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenString, 0, 1, 0, 2, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenNestedBraces(t *testing.T) {
	parser := NewParser("{foo{bar}}")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenString, 0, 1, 8, 10, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenEscapedBraces(t *testing.T) {
	parser := NewParser("{foo\\{bar\\}}")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenString, 0, 1, 10, 12, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenQuoted(t *testing.T) {
	parser := NewParser("\"foobar\"")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenEscape, 0, 1, 6, 8, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenQuotedEscapes(t *testing.T) {
	parser := NewParser("\"f\\to;o\\\"b\\na\\rr\"")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenEscape, 0, 1, 15, 17, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenSeparators(t *testing.T) {
	parser := NewParser("foo bar")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenEscape, 4, 0, 2, 3, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseToken()
	result = parserResult{stateOK, tokenSeparator, 3, 3, 3, 4, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseToken()
	result = parserResult{stateOK, tokenEscape, 0, 4, 6, 7, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenQuotedVariable(t *testing.T) {
	parser := NewParser("\"foo $bar\"")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenEscape, 5, 1, 4, 5, true}
	validateParser(state, parser, result, t)
	state, _ = parser.parseToken()
	result = parserResult{stateOK, tokenVariable, 1, 6, 8, 9, true}
	validateParser(state, parser, result, t)
	state, _ = parser.parseToken()
	result = parserResult{stateOK, tokenEscape, 0, 9, 8, 10, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenNestedCommand(t *testing.T) {
	parser := NewParser("\"foo [bar]\"")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenEscape, 6, 1, 4, 5, true}
	validateParser(state, parser, result, t)
	state, _ = parser.parseToken()
	result = parserResult{stateOK, tokenCommand, 1, 6, 8, 10, true}
	validateParser(state, parser, result, t)
	state, _ = parser.parseToken()
	result = parserResult{stateOK, tokenEscape, 0, 10, 9, 11, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenTrailingComment(t *testing.T) {
	parser := NewParser("foo; # bar")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenEscape, 7, 0, 2, 3, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseToken()
	result = parserResult{stateOK, tokenEOL, 5, 3, 4, 5, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseToken()
	result = parserResult{stateOK, tokenEOF, 0, 3, 4, 10, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenLeadingSeparator(t *testing.T) {
	parser := NewParser("   foo")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenSeparator, 3, 0, 2, 3, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseToken()
	result = parserResult{stateOK, tokenEscape, 0, 3, 5, 6, false}
	validateParser(state, parser, result, t)
}
