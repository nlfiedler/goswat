//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

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
	verifyLexerResults0(t, c, expected)
}

// verifyLexerExprResults calls lexExpr() and checks that the resulting
// tokens match the expected results.
func verifyLexerExprResults(t *testing.T, input string, expected []expectedLexerResult) {
	c := lexExpr("unit", input)
	verifyLexerResults0(t, c, expected)
}

// verifyLexerResults0 takes the output of lex() and lexExpr() and
// compares the results with the expected results.
func verifyLexerResults0(t *testing.T, c chan token, expected []expectedLexerResult) {
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

func TestLexerSetCommand(t *testing.T) {
	input := "set foo bar"
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenString, "set"})
	expected = append(expected, expectedLexerResult{tokenString, "foo"})
	expected = append(expected, expectedLexerResult{tokenString, "bar"})
	expected = append(expected, expectedLexerResult{tokenEOF, ""})
	verifyLexerResults(t, input, expected)
}

func TestLexerComments(t *testing.T) {
	input := `# foo
# bar baz
# quux
`
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenEOF, ""})
	verifyLexerResults(t, input, expected)
}

func TestLexerMixed(t *testing.T) {
	input := `# this is a test
[command foo bar]
$$x
$foo bar baz
[puts {hey [diddle] diddle} foo]
foobar
{}
{foo {bar}}
"foo [bar]"
"foo $bar"
"f\to;o\\\"b\na\rr"
`
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenCommand, "[command foo bar]"})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenString, "$"})
	expected = append(expected, expectedLexerResult{tokenVariable, "$x"})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenVariable, "$foo"})
	expected = append(expected, expectedLexerResult{tokenString, "bar"})
	expected = append(expected, expectedLexerResult{tokenString, "baz"})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenCommand, "[puts {hey [diddle] diddle} foo]"})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenString, "foobar"})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenBrace, "{}"})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenBrace, "{foo {bar}}"})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenQuote, "\"foo "})
	expected = append(expected, expectedLexerResult{tokenCommand, "[bar]"})
	expected = append(expected, expectedLexerResult{tokenQuote, "\""})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenQuote, "\"foo "})
	expected = append(expected, expectedLexerResult{tokenVariable, "$bar"})
	expected = append(expected, expectedLexerResult{tokenQuote, "\""})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenQuote, `"f\to;o\\\"b\na\rr"`})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	verifyLexerResults(t, input, expected)
}

func TestLexerGroupingBraces(t *testing.T) {
	input := `set Z Albany
set Z_LABEL "The Capitol of New York is: "

puts "\n................. examples of differences between  \" and \{"
puts "$Z_LABEL $Z"
puts {$Z_LABEL $Z}

puts "\n....... examples of differences in nesting \{ and \" "
puts "$Z_LABEL {$Z}"
puts {Who said, "What this country needs is a good $0.05 cigar!"?}

puts "\n................. examples of escape strings"
puts {There are no substitutions done within braces \n \r \x0a \f \v}
puts {But, the escaped newline at the end of a\
string is still evaluated as a space}
`
	expected := make([]expectedLexerResult, 0)
	// 0
	expected = append(expected, expectedLexerResult{tokenString, "set"})
	expected = append(expected, expectedLexerResult{tokenString, "Z"})
	expected = append(expected, expectedLexerResult{tokenString, "Albany"})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenString, "set"})
	// 5
	expected = append(expected, expectedLexerResult{tokenString, "Z_LABEL"})
	expected = append(expected, expectedLexerResult{tokenQuote, `"The Capitol of New York is: "`})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n\n"})
	expected = append(expected, expectedLexerResult{tokenString, "puts"})
	expected = append(expected, expectedLexerResult{tokenQuote, `"\n................. examples of differences between  \" and \{"`})
	// 10
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenString, "puts"})
	expected = append(expected, expectedLexerResult{tokenQuote, `"`})
	expected = append(expected, expectedLexerResult{tokenVariable, `$Z_LABEL`})
	expected = append(expected, expectedLexerResult{tokenQuote, " "})
	// 15
	expected = append(expected, expectedLexerResult{tokenVariable, `$Z`})
	expected = append(expected, expectedLexerResult{tokenQuote, `"`})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenString, "puts"})
	expected = append(expected, expectedLexerResult{tokenBrace, `{$Z_LABEL $Z}`})
	// 20
	expected = append(expected, expectedLexerResult{tokenEOL, "\n\n"})
	expected = append(expected, expectedLexerResult{tokenString, "puts"})
	expected = append(expected, expectedLexerResult{tokenQuote, `"\n....... examples of differences in nesting \{ and \" "`})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenString, "puts"})
	// 25
	expected = append(expected, expectedLexerResult{tokenQuote, `"`})
	expected = append(expected, expectedLexerResult{tokenVariable, `$Z_LABEL`})
	expected = append(expected, expectedLexerResult{tokenQuote, " {"})
	expected = append(expected, expectedLexerResult{tokenVariable, `$Z`})
	expected = append(expected, expectedLexerResult{tokenQuote, `}"`})
	// 30
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenString, "puts"})
	expected = append(expected, expectedLexerResult{tokenBrace, `{Who said, "What this country needs is a good $0.05 cigar!"?}`})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n\n"})
	expected = append(expected, expectedLexerResult{tokenString, "puts"})
	// 35
	expected = append(expected, expectedLexerResult{tokenQuote, `"\n................. examples of escape strings"`})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenString, "puts"})
	expected = append(expected, expectedLexerResult{tokenBrace, `{There are no substitutions done within braces \n \r \x0a \f \v}`})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	// 40
	expected = append(expected, expectedLexerResult{tokenString, "puts"})
	expected = append(expected, expectedLexerResult{tokenBrace, `{But, the escaped newline at the end of a\
string is still evaluated as a space}`}) // TODO: add test to interpreter_test.go, expect \n to be replaced with space
	verifyLexerResults(t, input, expected)
}

func TestLexerCommands(t *testing.T) {
	input := `set y [set x "def"]
set z {[set x "This is a string within quotes within braces"]}
set a "[set x {This is a string within braces within quotes}]"
set b "\[set y {This is a string within braces within quotes}]"
`
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenString, "set"})
	expected = append(expected, expectedLexerResult{tokenString, "y"})
	expected = append(expected, expectedLexerResult{tokenCommand, `[set x "def"]`})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenString, "set"})
	expected = append(expected, expectedLexerResult{tokenString, "z"})
	expected = append(expected, expectedLexerResult{tokenBrace, `{[set x "This is a string within quotes within braces"]}`})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenString, "set"})
	expected = append(expected, expectedLexerResult{tokenString, "a"})
	expected = append(expected, expectedLexerResult{tokenQuote, `"`})
	expected = append(expected, expectedLexerResult{tokenCommand, "[set x {This is a string within braces within quotes}]"})
	expected = append(expected, expectedLexerResult{tokenQuote, `"`})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	expected = append(expected, expectedLexerResult{tokenString, "set"})
	expected = append(expected, expectedLexerResult{tokenString, "b"})
	expected = append(expected, expectedLexerResult{tokenQuote, `"\[set y {This is a string within braces within quotes}]"`})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	verifyLexerResults(t, input, expected)
}

func TestLexerForLoop(t *testing.T) {
	input := `for { set i 0 } { $i <= $number } { incr i } {
   set x [expr {$i*0.1}]
   create label $x
}
`
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenString, "for"})
	expected = append(expected, expectedLexerResult{tokenBrace, "{ set i 0 }"})
	expected = append(expected, expectedLexerResult{tokenBrace, "{ $i <= $number }"})
	expected = append(expected, expectedLexerResult{tokenBrace, "{ incr i }"})
	expected = append(expected, expectedLexerResult{tokenBrace, "{\n   set x [expr {$i*0.1}]\n   create label $x\n}"})
	expected = append(expected, expectedLexerResult{tokenEOL, "\n"})
	verifyLexerResults(t, input, expected)
}

func TestLexerIfElse(t *testing.T) {
	input := `if {$x != 1} {
    puts "$x is != 1"
} else {
    puts "$x is 1"
}`
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenString, "if"})
	expected = append(expected, expectedLexerResult{tokenBrace, "{$x != 1}"})
	expected = append(expected, expectedLexerResult{tokenBrace, "{\n    puts \"$x is != 1\"\n}"})
	expected = append(expected, expectedLexerResult{tokenString, "else"})
	expected = append(expected, expectedLexerResult{tokenBrace, "{\n    puts \"$x is 1\"\n}"})
	verifyLexerResults(t, input, expected)
}

func TestLexerUnclosedQuotes(t *testing.T) {
	input := make(map[string]expectedLexerError)
	input[`"foo`] = expectedLexerError{"unclosed quoted string", "unclosed quote should fail"}
	input[`{foo`] = expectedLexerError{"unclosed left brace", "unclosed brace should fail"}
	input[`${foo`] = expectedLexerError{"unclosed variable", "unclosed variable should fail"}
	input[`[foo`] = expectedLexerError{"unclosed command", "unclosed command should fail"}
	verifyLexerErrors(t, input)
}

func TestLexerNumbers(t *testing.T) {
	input := "1 2.1 3. 6E4 7.91e+16 .000001 0366 0x7b5"
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenInteger, "1"})
	expected = append(expected, expectedLexerResult{tokenFloat, "2.1"})
	expected = append(expected, expectedLexerResult{tokenFloat, "3."})
	expected = append(expected, expectedLexerResult{tokenFloat, "6E4"})
	expected = append(expected, expectedLexerResult{tokenFloat, "7.91e+16"})
	expected = append(expected, expectedLexerResult{tokenFloat, ".000001"})
	expected = append(expected, expectedLexerResult{tokenInteger, "0366"})
	expected = append(expected, expectedLexerResult{tokenInteger, "0x7b5"})
	verifyLexerExprResults(t, input, expected)
}

func TestLexerOperators(t *testing.T) {
	input := "- + ~ ! * / % < > = & ^ | ? : ** << >> <= >= && ||"
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenOperator, "-"})
	expected = append(expected, expectedLexerResult{tokenOperator, "+"})
	expected = append(expected, expectedLexerResult{tokenOperator, "~"})
	expected = append(expected, expectedLexerResult{tokenOperator, "!"})
	expected = append(expected, expectedLexerResult{tokenOperator, "*"})
	expected = append(expected, expectedLexerResult{tokenOperator, "/"})
	expected = append(expected, expectedLexerResult{tokenOperator, "%"})
	expected = append(expected, expectedLexerResult{tokenOperator, "<"})
	expected = append(expected, expectedLexerResult{tokenOperator, ">"})
	expected = append(expected, expectedLexerResult{tokenOperator, "="})
	expected = append(expected, expectedLexerResult{tokenOperator, "&"})
	expected = append(expected, expectedLexerResult{tokenOperator, "^"})
	expected = append(expected, expectedLexerResult{tokenOperator, "|"})
	expected = append(expected, expectedLexerResult{tokenOperator, "?"})
	expected = append(expected, expectedLexerResult{tokenOperator, ":"})
	expected = append(expected, expectedLexerResult{tokenOperator, "**"})
	expected = append(expected, expectedLexerResult{tokenOperator, "<<"})
	expected = append(expected, expectedLexerResult{tokenOperator, ">>"})
	expected = append(expected, expectedLexerResult{tokenOperator, "<="})
	expected = append(expected, expectedLexerResult{tokenOperator, ">="})
	expected = append(expected, expectedLexerResult{tokenOperator, "&&"})
	expected = append(expected, expectedLexerResult{tokenOperator, "||"})
	verifyLexerExprResults(t, input, expected)
}

func TestLexerFunctions(t *testing.T) {
	input := "a12() abs(123) max(1, 2) foo"
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenFunction, "a12("})
	expected = append(expected, expectedLexerResult{tokenParen, ")"})
	expected = append(expected, expectedLexerResult{tokenFunction, "abs("})
	expected = append(expected, expectedLexerResult{tokenInteger, "123"})
	expected = append(expected, expectedLexerResult{tokenParen, ")"})
	expected = append(expected, expectedLexerResult{tokenFunction, "max("})
	expected = append(expected, expectedLexerResult{tokenInteger, "1"})
	expected = append(expected, expectedLexerResult{tokenComma, ","})
	expected = append(expected, expectedLexerResult{tokenInteger, "2"})
	expected = append(expected, expectedLexerResult{tokenParen, ")"})
	expected = append(expected, expectedLexerResult{tokenError, "apparent function call missing (: \"foo\""})
	verifyLexerExprResults(t, input, expected)
}

func TestLexerExprNewline(t *testing.T) {
	input := `123
$foo`
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenInteger, "123"})
	expected = append(expected, expectedLexerResult{tokenError, "newline, semicolon, and hash not allowed in expression"})
	verifyLexerExprResults(t, input, expected)
}

func TestLexerExprComment(t *testing.T) {
	input := "123 # foo"
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenInteger, "123"})
	expected = append(expected, expectedLexerResult{tokenError, "newline, semicolon, and hash not allowed in expression"})
	verifyLexerExprResults(t, input, expected)
}

func TestLexerExprSemicolon(t *testing.T) {
	input := "123 ; foo"
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenInteger, "123"})
	expected = append(expected, expectedLexerResult{tokenError, "newline, semicolon, and hash not allowed in expression"})
	verifyLexerExprResults(t, input, expected)
}

func TestLexerTokenConents(t *testing.T) {
	input := `$foo ${foo} [foo] "foo [bar] baz quux" {foo bar} foobar`
	expected := make([]expectedLexerResult, 0)
	expected = append(expected, expectedLexerResult{tokenVariable, "foo"})
	expected = append(expected, expectedLexerResult{tokenVariable, "foo"})
	expected = append(expected, expectedLexerResult{tokenCommand, "foo"})
	expected = append(expected, expectedLexerResult{tokenQuote, "foo "})
	expected = append(expected, expectedLexerResult{tokenCommand, "bar"})
	expected = append(expected, expectedLexerResult{tokenQuote, " baz quux"})
	expected = append(expected, expectedLexerResult{tokenBrace, "foo bar"})
	expected = append(expected, expectedLexerResult{tokenString, "foobar"})
	c := lex("unit", input)
	for i, e := range expected {
		token, ok := <-c
		if !ok {
			t.Errorf("lexer channel closed?")
		}
		if token.typ != e.typ {
			t.Errorf("expected %d, got %d for '%s' (token %d)", e.typ, token.typ, e.val, i)
		}
		txt := token.contents()
		if txt != e.val {
			t.Errorf("expected '%s', got '%s' (token %d, type %d)", e.val, txt, i, e.typ)
		}
	}
	drainLexerChannel(c)
}
