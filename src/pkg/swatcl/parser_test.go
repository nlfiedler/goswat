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
	len   int
	start int
	end   int
	p     int
	quote bool
}

// validateParser checks that the parser is in exactly the state
// specified by the result, failing if that is not the case.
func validateParser(state parserState, parser *Parser, result parserResult, t *testing.T) {
	if state != result.state {
		t.Errorf("state did not match for %s (expected %d, actual %d)",
			parser.text, result.state, state)
	}
	if parser.token != result.token {
		t.Errorf("token did not match for %s (expected %d, actual %d)",
			parser.text, result.token, parser.token)
	}
	if parser.start != result.start {
		t.Errorf("start did not match for %s (expected %d, actual %d)",
			parser.text, result.start, parser.start)
	}
	if parser.p != result.p {
		t.Errorf("p did not match for %s (expected %d, actual %d)",
			parser.text, result.p, parser.p)
	}
	if parser.end != result.end {
		t.Errorf("end did not match for %s (expected %d, actual %d)",
			parser.text, result.end, parser.end)
	}
	if parser.len != result.len {
		t.Errorf("len did not match for %s (expected %d, actual %d)",
			parser.text, result.len, parser.len)
	}
	if parser.insidequote != result.quote {
		t.Errorf("insidequote did not match for %s (expected %t, actual %t)",
			parser.text, result.quote, parser.insidequote)
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

func TestParseVariableDoubleDollar(t *testing.T) {
	parser := NewParser("$$a")
	state, _ := parser.parseVariable()
	result := parserResult{stateOK, tokenString, 2, 0, 0, 1, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseVariable()
	result = parserResult{stateOK, tokenVariable, 0, 2, 2, 3, false}
	validateParser(state, parser, result, t)
}

//
// parseBrace
//

func TestParseBrace(t *testing.T) {
	parser := NewParser("{foo}")
	state, _ := parser.parseBrace()
	result := parserResult{stateOK, tokenBrace, 0, 1, 3, 5, false}
	validateParser(state, parser, result, t)
}

func TestParseBraceEmpty(t *testing.T) {
	parser := NewParser("{}")
	state, _ := parser.parseBrace()
	result := parserResult{stateOK, tokenBrace, 0, 1, 0, 2, false}
	validateParser(state, parser, result, t)
}

func TestParseBraceNested(t *testing.T) {
	parser := NewParser("{foo{bar}}")
	state, _ := parser.parseBrace()
	result := parserResult{stateOK, tokenBrace, 0, 1, 8, 10, false}
	validateParser(state, parser, result, t)
}

func TestParseBraceEscaped(t *testing.T) {
	parser := NewParser("{foo\\{bar\\}}")
	state, _ := parser.parseBrace()
	result := parserResult{stateOK, tokenBrace, 0, 1, 10, 12, false}
	validateParser(state, parser, result, t)
}

//
// parseString
//

func TestParseStringEmptyBraces(t *testing.T) {
	parser := NewParser("{}")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenBrace, 0, 1, 0, 2, false}
	validateParser(state, parser, result, t)
}

func TestParseStringNestedBraces(t *testing.T) {
	parser := NewParser("{foo{bar}}")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenBrace, 0, 1, 8, 10, false}
	validateParser(state, parser, result, t)
}

func TestParseStringEscapedBraces(t *testing.T) {
	parser := NewParser("{foo\\{bar\\}}")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenBrace, 0, 1, 10, 12, false}
	validateParser(state, parser, result, t)
}

func TestParseString(t *testing.T) {
	parser := NewParser("\"foobar\"")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenString, 0, 1, 6, 8, false}
	validateParser(state, parser, result, t)
}

func TestParseStringEscapes(t *testing.T) {
	parser := NewParser("\"f\\to;o\\\"b\\na\\rr\"")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenString, 0, 1, 15, 17, false}
	validateParser(state, parser, result, t)
}

func TestParseStringSeparators(t *testing.T) {
	parser := NewParser("foo bar")
	state, _ := parser.parseString()
	result := parserResult{stateOK, tokenString, 4, 0, 2, 3, false}
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
	result := parserResult{stateOK, tokenString, 0, 0, 5, 6, false}
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
	result := parserResult{stateOK, tokenString, 0, 0, 5, 6, false}
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
	result := parserResult{stateOK, tokenBrace, 0, 1, 0, 2, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenNestedBraces(t *testing.T) {
	parser := NewParser("{foo{bar}}")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenBrace, 0, 1, 8, 10, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenEscapedBraces(t *testing.T) {
	parser := NewParser("{foo\\{bar\\}}")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenBrace, 0, 1, 10, 12, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenQuoted(t *testing.T) {
	parser := NewParser("\"foobar\"")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenString, 0, 1, 6, 8, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenQuotedEscapes(t *testing.T) {
	parser := NewParser("\"f\\to;o\\\"b\\na\\rr\"")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenString, 0, 1, 15, 17, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenSeparators(t *testing.T) {
	parser := NewParser("foo bar")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenString, 4, 0, 2, 3, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseToken()
	result = parserResult{stateOK, tokenSeparator, 3, 3, 3, 4, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseToken()
	result = parserResult{stateOK, tokenString, 0, 4, 6, 7, false}
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
	result = parserResult{stateOK, tokenString, 0, 9, 8, 10, false}
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
	result = parserResult{stateOK, tokenString, 0, 10, 9, 11, false}
	validateParser(state, parser, result, t)
}

func TestParseTokenTrailingComment(t *testing.T) {
	parser := NewParser("foo; # bar")
	state, _ := parser.parseToken()
	result := parserResult{stateOK, tokenString, 7, 0, 2, 3, false}
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
	result = parserResult{stateOK, tokenString, 0, 3, 5, 6, false}
	validateParser(state, parser, result, t)
}

//
// parseOperator
//

func TestParseOperator(t *testing.T) {
	parser := NewParser("*a")
	state, _ := parser.parseOperator()
	result := parserResult{stateOK, tokenOperator, 1, 0, 0, 1, false}
	validateParser(state, parser, result, t)

	parser = NewParser("**")
	state, _ = parser.parseOperator()
	result = parserResult{stateOK, tokenOperator, 0, 0, 1, 2, false}
	validateParser(state, parser, result, t)

	parser = NewParser("*")
	state, _ = parser.parseOperator()
	result = parserResult{stateOK, tokenOperator, 0, 0, 0, 1, false}
	validateParser(state, parser, result, t)

	parser = NewParser("&")
	state, _ = parser.parseOperator()
	result = parserResult{stateOK, tokenOperator, 0, 0, 0, 1, false}
	validateParser(state, parser, result, t)

	parser = NewParser("!")
	state, _ = parser.parseOperator()
	result = parserResult{stateOK, tokenOperator, 0, 0, 0, 1, false}
	validateParser(state, parser, result, t)

	parser = NewParser("<=")
	state, _ = parser.parseOperator()
	result = parserResult{stateOK, tokenOperator, 0, 0, 1, 2, false}
	validateParser(state, parser, result, t)

	parser = NewParser("&&")
	state, _ = parser.parseOperator()
	result = parserResult{stateOK, tokenOperator, 0, 0, 1, 2, false}
	validateParser(state, parser, result, t)

	parser = NewParser("||")
	state, _ = parser.parseOperator()
	result = parserResult{stateOK, tokenOperator, 0, 0, 1, 2, false}
	validateParser(state, parser, result, t)
}

//
// parseNumber
//

func TestParseNumber(t *testing.T) {
	parser := NewParser("0x00bab10c")
	state, _ := parser.parseNumber()
	result := parserResult{stateOK, tokenInteger, 0, 0, 9, 10, false}
	validateParser(state, parser, result, t)

	parser = NewParser("1")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenInteger, 0, 0, 0, 1, false}
	validateParser(state, parser, result, t)

	parser = NewParser("2.1")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenFloat, 0, 0, 2, 3, false}
	validateParser(state, parser, result, t)

	parser = NewParser("3.")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenFloat, 0, 0, 1, 2, false}
	validateParser(state, parser, result, t)

	parser = NewParser(".0001")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenFloat, 0, 0, 4, 5, false}
	validateParser(state, parser, result, t)

	parser = NewParser("6E4")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenFloat, 0, 0, 2, 3, false}
	validateParser(state, parser, result, t)

	parser = NewParser("6+0")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenInteger, 2, 0, 0, 1, false}
	validateParser(state, parser, result, t)

	parser = NewParser("6E")
	state, _ = parser.parseNumber()
	result = parserResult{stateError, tokenEOL, 0, 0, 0, 2, false}
	validateParser(state, parser, result, t)

	parser = NewParser("7.91e+16")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenFloat, 0, 0, 7, 8, false}
	validateParser(state, parser, result, t)

	parser = NewParser("1e+012")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenFloat, 0, 0, 5, 6, false}
	validateParser(state, parser, result, t)

	parser = NewParser("0366")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenInteger, 0, 0, 3, 4, false}
	validateParser(state, parser, result, t)

	parser = NewParser("0070/")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenInteger, 1, 0, 3, 4, false}
	validateParser(state, parser, result, t)

	parser = NewParser("0070*")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenInteger, 1, 0, 3, 4, false}
	validateParser(state, parser, result, t)

	parser = NewParser("10101010")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenInteger, 0, 0, 7, 8, false}
	validateParser(state, parser, result, t)

	parser = NewParser("4.4408920985006262e-016")
	state, _ = parser.parseNumber()
	result = parserResult{stateOK, tokenFloat, 0, 0, 22, 23, false}
	validateParser(state, parser, result, t)

	parser = NewParser("+123")
	state, _ = parser.parseNumber()
	result = parserResult{stateError, tokenEOL, 4, 0, 0, 0, false}
	validateParser(state, parser, result, t)

	parser = NewParser("-42")
	state, _ = parser.parseNumber()
	result = parserResult{stateError, tokenEOL, 3, 0, 0, 0, false}
	validateParser(state, parser, result, t)

	parser = NewParser("!@$%")
	state, _ = parser.parseNumber()
	result = parserResult{stateError, tokenEOL, 4, 0, 0, 0, false}
	validateParser(state, parser, result, t)

	parser = NewParser("a.10")
	state, _ = parser.parseNumber()
	result = parserResult{stateError, tokenEOL, 4, 0, 0, 0, false}
	validateParser(state, parser, result, t)
}

//
// parseFunction
//

func TestParseFunction(t *testing.T) {
	parser := NewParser("atoi('123')")
	state, _ := parser.parseFunction()
	result := parserResult{stateOK, tokenFunction, 6, 0, 4, 5, false}
	validateParser(state, parser, result, t)

	parser = NewParser("AT0I('123')")
	state, _ = parser.parseFunction()
	result = parserResult{stateOK, tokenFunction, 6, 0, 4, 5, false}
	validateParser(state, parser, result, t)
}

//
// parseExprToken
//

func TestParseExprTokenBlank(t *testing.T) {
	parser := NewParser("")
	state, _ := parser.parseExprToken()
	result := parserResult{stateOK, tokenEOF, 0, 0, 0, 0, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenWord(t *testing.T) {
	parser := NewParser("foobar")
	state, _ := parser.parseExprToken()
	result := parserResult{stateError, tokenEOL, 0, 0, 0, 6, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenVariable(t *testing.T) {
	parser := NewParser("$foobar")
	state, _ := parser.parseExprToken()
	result := parserResult{stateOK, tokenVariable, 0, 1, 6, 7, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenEol(t *testing.T) {
	parser := NewParser("; \n \t \r")
	state, _ := parser.parseExprToken()
	result := parserResult{stateError, tokenEOL, 7, 0, 0, 0, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenComment(t *testing.T) {
	parser := NewParser("# foo")
	state, _ := parser.parseExprToken()
	result := parserResult{stateError, tokenEOL, 5, 0, 0, 0, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenEmptyBraces(t *testing.T) {
	parser := NewParser("{}")
	state, _ := parser.parseExprToken()
	result := parserResult{stateOK, tokenBrace, 0, 1, 0, 2, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenNestedBraces(t *testing.T) {
	parser := NewParser("{foo{bar}}")
	state, _ := parser.parseExprToken()
	result := parserResult{stateOK, tokenBrace, 0, 1, 8, 10, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenEscapedBraces(t *testing.T) {
	parser := NewParser("{foo\\{bar\\}}")
	state, _ := parser.parseExprToken()
	result := parserResult{stateOK, tokenBrace, 0, 1, 10, 12, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenQuoted(t *testing.T) {
	parser := NewParser("\"foobar\"")
	state, _ := parser.parseExprToken()
	result := parserResult{stateOK, tokenString, 0, 1, 6, 8, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenQuotedEscapes(t *testing.T) {
	parser := NewParser("\"f\\to;o\\\"b\\na\\rr\"")
	state, _ := parser.parseExprToken()
	result := parserResult{stateOK, tokenString, 0, 1, 15, 17, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenSeparators(t *testing.T) {
	parser := NewParser("$foo + $bar")
	state, _ := parser.parseExprToken()
	result := parserResult{stateOK, tokenVariable, 7, 1, 3, 4, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseExprToken()
	result = parserResult{stateOK, tokenSeparator, 6, 4, 4, 5, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseExprToken()
	result = parserResult{stateOK, tokenOperator, 5, 5, 5, 6, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseExprToken()
	result = parserResult{stateOK, tokenSeparator, 4, 6, 6, 7, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseExprToken()
	result = parserResult{stateOK, tokenVariable, 0, 8, 10, 11, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenQuotedVariable(t *testing.T) {
	parser := NewParser("\"foo $bar\"")
	state, _ := parser.parseExprToken()
	result := parserResult{stateOK, tokenEscape, 5, 1, 4, 5, true}
	validateParser(state, parser, result, t)
	state, _ = parser.parseExprToken()
	result = parserResult{stateOK, tokenVariable, 1, 6, 8, 9, true}
	validateParser(state, parser, result, t)
	state, _ = parser.parseExprToken()
	result = parserResult{stateOK, tokenString, 0, 9, 8, 10, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenNestedCommand(t *testing.T) {
	parser := NewParser("\"foo [bar]\"")
	state, _ := parser.parseExprToken()
	result := parserResult{stateOK, tokenEscape, 6, 1, 4, 5, true}
	validateParser(state, parser, result, t)
	state, _ = parser.parseExprToken()
	result = parserResult{stateOK, tokenCommand, 1, 6, 8, 10, true}
	validateParser(state, parser, result, t)
	state, _ = parser.parseExprToken()
	result = parserResult{stateOK, tokenString, 0, 10, 9, 11, false}
	validateParser(state, parser, result, t)
}

func TestParseExprTokenLeadingSeparator(t *testing.T) {
	parser := NewParser("   $foo")
	state, _ := parser.parseExprToken()
	result := parserResult{stateOK, tokenSeparator, 4, 0, 2, 3, false}
	validateParser(state, parser, result, t)
	state, _ = parser.parseExprToken()
	result = parserResult{stateOK, tokenVariable, 0, 4, 6, 7, false}
	validateParser(state, parser, result, t)
}

// TODO: write parseExprToken tests for parsing function calls
