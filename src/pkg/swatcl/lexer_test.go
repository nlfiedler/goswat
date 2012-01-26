//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

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
	verifyLexerResults0(t, c, expected)
}

// verifyLexerExprResults calls lexExpr() and checks that the resulting
// tokens match the expected results.
func verifyLexerExprResults(t *testing.T, input string, expected []expectedResult) {
	c := lexExpr("unit", input)
	verifyLexerResults0(t, c, expected)
}

// verifyLexerResults0 takes the output of lex() and lexExpr() and
// compares the results with the expected results.
func verifyLexerResults0(t *testing.T, c chan token, expected []expectedResult) {
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

func TestLexerSetCommand(t *testing.T) {
	input := "set foo bar"
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenString, "set"})
	expected = append(expected, expectedResult{tokenString, "foo"})
	expected = append(expected, expectedResult{tokenString, "bar"})
	expected = append(expected, expectedResult{tokenEOF, ""})
	verifyLexerResults(t, input, expected)
}

func TestLexerComments(t *testing.T) {
	input := `# foo
# bar baz
# quux
`
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenEOF, ""})
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
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenCommand, "[command foo bar]"})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenString, "$"})
	expected = append(expected, expectedResult{tokenVariable, "$x"})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenVariable, "$foo"})
	expected = append(expected, expectedResult{tokenString, "bar"})
	expected = append(expected, expectedResult{tokenString, "baz"})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenCommand, "[puts {hey [diddle] diddle} foo]"})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenString, "foobar"})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenBrace, "{}"})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenBrace, "{foo {bar}}"})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenQuote, "\"foo "})
	expected = append(expected, expectedResult{tokenCommand, "[bar]"})
	expected = append(expected, expectedResult{tokenQuote, "\""})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenQuote, "\"foo "})
	expected = append(expected, expectedResult{tokenVariable, "$bar"})
	expected = append(expected, expectedResult{tokenQuote, "\""})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenQuote, `"f\to;o\\\"b\na\rr"`})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
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
	expected := make([]expectedResult, 0)
	// 0
	expected = append(expected, expectedResult{tokenString, "set"})
	expected = append(expected, expectedResult{tokenString, "Z"})
	expected = append(expected, expectedResult{tokenString, "Albany"})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenString, "set"})
	// 5
	expected = append(expected, expectedResult{tokenString, "Z_LABEL"})
	expected = append(expected, expectedResult{tokenQuote, `"The Capitol of New York is: "`})
	expected = append(expected, expectedResult{tokenEOL, "\n\n"})
	expected = append(expected, expectedResult{tokenString, "puts"})
	expected = append(expected, expectedResult{tokenQuote, `"\n................. examples of differences between  \" and \{"`})
	// 10
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenString, "puts"})
	expected = append(expected, expectedResult{tokenQuote, `"`})
	expected = append(expected, expectedResult{tokenVariable, `$Z_LABEL`})
	expected = append(expected, expectedResult{tokenQuote, " "})
	// 15
	expected = append(expected, expectedResult{tokenVariable, `$Z`})
	expected = append(expected, expectedResult{tokenQuote, `"`})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenString, "puts"})
	expected = append(expected, expectedResult{tokenBrace, `{$Z_LABEL $Z}`})
	// 20
	expected = append(expected, expectedResult{tokenEOL, "\n\n"})
	expected = append(expected, expectedResult{tokenString, "puts"})
	expected = append(expected, expectedResult{tokenQuote, `"\n....... examples of differences in nesting \{ and \" "`})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenString, "puts"})
	// 25
	expected = append(expected, expectedResult{tokenQuote, `"`})
	expected = append(expected, expectedResult{tokenVariable, `$Z_LABEL`})
	expected = append(expected, expectedResult{tokenQuote, " {"})
	expected = append(expected, expectedResult{tokenVariable, `$Z`})
	expected = append(expected, expectedResult{tokenQuote, `}"`})
	// 30
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenString, "puts"})
	expected = append(expected, expectedResult{tokenBrace, `{Who said, "What this country needs is a good $0.05 cigar!"?}`})
	expected = append(expected, expectedResult{tokenEOL, "\n\n"})
	expected = append(expected, expectedResult{tokenString, "puts"})
	// 35
	expected = append(expected, expectedResult{tokenQuote, `"\n................. examples of escape strings"`})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenString, "puts"})
	expected = append(expected, expectedResult{tokenBrace, `{There are no substitutions done within braces \n \r \x0a \f \v}`})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	// 40
	expected = append(expected, expectedResult{tokenString, "puts"})
	expected = append(expected, expectedResult{tokenBrace, `{But, the escaped newline at the end of a\
string is still evaluated as a space}`}) // TODO: add test to interpreter_test.go, expect \n to be replaced with space
	verifyLexerResults(t, input, expected)
}

func TestLexerCommands(t *testing.T) {
	input := `set y [set x "def"]
set z {[set x "This is a string within quotes within braces"]}
set a "[set x {This is a string within braces within quotes}]"
set b "\[set y {This is a string within braces within quotes}]"
`
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenString, "set"})
	expected = append(expected, expectedResult{tokenString, "y"})
	expected = append(expected, expectedResult{tokenCommand, `[set x "def"]`})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenString, "set"})
	expected = append(expected, expectedResult{tokenString, "z"})
	expected = append(expected, expectedResult{tokenBrace, `{[set x "This is a string within quotes within braces"]}`})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenString, "set"})
	expected = append(expected, expectedResult{tokenString, "a"})
	expected = append(expected, expectedResult{tokenQuote, `"`})
	expected = append(expected, expectedResult{tokenCommand, "[set x {This is a string within braces within quotes}]"})
	expected = append(expected, expectedResult{tokenQuote, `"`})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	expected = append(expected, expectedResult{tokenString, "set"})
	expected = append(expected, expectedResult{tokenString, "b"})
	expected = append(expected, expectedResult{tokenQuote, `"\[set y {This is a string within braces within quotes}]"`})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	verifyLexerResults(t, input, expected)
}

func TestLexerForLoop(t *testing.T) {
	input := `for { set i 0 } { $i <= $number } { incr i } {
   set x [expr {$i*0.1}]
   create label $x
}
`
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenString, "for"})
	expected = append(expected, expectedResult{tokenBrace, "{ set i 0 }"})
	expected = append(expected, expectedResult{tokenBrace, "{ $i <= $number }"})
	expected = append(expected, expectedResult{tokenBrace, "{ incr i }"})
	expected = append(expected, expectedResult{tokenBrace, "{\n   set x [expr {$i*0.1}]\n   create label $x\n}"})
	expected = append(expected, expectedResult{tokenEOL, "\n"})
	verifyLexerResults(t, input, expected)
}

func TestLexerIfElse(t *testing.T) {
	input := `if {$x != 1} {
    puts "$x is != 1"
} else {
    puts "$x is 1"
}`
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenString, "if"})
	expected = append(expected, expectedResult{tokenBrace, "{$x != 1}"})
	expected = append(expected, expectedResult{tokenBrace, "{\n    puts \"$x is != 1\"\n}"})
	expected = append(expected, expectedResult{tokenString, "else"})
	expected = append(expected, expectedResult{tokenBrace, "{\n    puts \"$x is 1\"\n}"})
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

func TestLexerUnclosedVariable(t *testing.T) {
	input := `${foo`
	c := lex("unit", input)
	token, ok := <-c
	if !ok {
		t.Errorf("lexer channel closed?")
	}
	if token.typ != tokenError {
		t.Errorf("expected lexing unclosed variable to fail")
	}
}

func TestLexerUnclosedBrace(t *testing.T) {
	input := `{foo`
	c := lex("unit", input)
	token, ok := <-c
	if !ok {
		t.Errorf("lexer channel closed?")
	}
	if token.typ != tokenError {
		t.Errorf("expected lexing unclosed brace to fail")
	}
}

func TestLexerUnclosedCommand(t *testing.T) {
	input := `[foo`
	c := lex("unit", input)
	token, ok := <-c
	if !ok {
		t.Errorf("lexer channel closed?")
	}
	if token.typ != tokenError {
		t.Errorf("expected lexing unclosed command to fail")
	}
}

func TestLexerNumbers(t *testing.T) {
	input := "1 2.1 3. 6E4 7.91e+16 .000001 0366 0x7b5"
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenInteger, "1"})
	expected = append(expected, expectedResult{tokenFloat, "2.1"})
	expected = append(expected, expectedResult{tokenFloat, "3."})
	expected = append(expected, expectedResult{tokenFloat, "6E4"})
	expected = append(expected, expectedResult{tokenFloat, "7.91e+16"})
	expected = append(expected, expectedResult{tokenFloat, ".000001"})
	expected = append(expected, expectedResult{tokenInteger, "0366"})
	expected = append(expected, expectedResult{tokenInteger, "0x7b5"})
	verifyLexerExprResults(t, input, expected)
}

func TestLexerOperators(t *testing.T) {
	input := "- + ~ ! * / % < > = & ^ | ? : ** << >> <= >= && ||"
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenOperator, "-"})
	expected = append(expected, expectedResult{tokenOperator, "+"})
	expected = append(expected, expectedResult{tokenOperator, "~"})
	expected = append(expected, expectedResult{tokenOperator, "!"})
	expected = append(expected, expectedResult{tokenOperator, "*"})
	expected = append(expected, expectedResult{tokenOperator, "/"})
	expected = append(expected, expectedResult{tokenOperator, "%"})
	expected = append(expected, expectedResult{tokenOperator, "<"})
	expected = append(expected, expectedResult{tokenOperator, ">"})
	expected = append(expected, expectedResult{tokenOperator, "="})
	expected = append(expected, expectedResult{tokenOperator, "&"})
	expected = append(expected, expectedResult{tokenOperator, "^"})
	expected = append(expected, expectedResult{tokenOperator, "|"})
	expected = append(expected, expectedResult{tokenOperator, "?"})
	expected = append(expected, expectedResult{tokenOperator, ":"})
	expected = append(expected, expectedResult{tokenOperator, "**"})
	expected = append(expected, expectedResult{tokenOperator, "<<"})
	expected = append(expected, expectedResult{tokenOperator, ">>"})
	expected = append(expected, expectedResult{tokenOperator, "<="})
	expected = append(expected, expectedResult{tokenOperator, ">="})
	expected = append(expected, expectedResult{tokenOperator, "&&"})
	expected = append(expected, expectedResult{tokenOperator, "||"})
	verifyLexerExprResults(t, input, expected)
}

func TestLexerFunctions(t *testing.T) {
	input := "a12() abs(123) max(1, 2) foo"
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenFunction, "a12("})
	expected = append(expected, expectedResult{tokenParen, ")"})
	expected = append(expected, expectedResult{tokenFunction, "abs("})
	expected = append(expected, expectedResult{tokenInteger, "123"})
	expected = append(expected, expectedResult{tokenParen, ")"})
	expected = append(expected, expectedResult{tokenFunction, "max("})
	expected = append(expected, expectedResult{tokenInteger, "1"})
	expected = append(expected, expectedResult{tokenComma, ","})
	expected = append(expected, expectedResult{tokenInteger, "2"})
	expected = append(expected, expectedResult{tokenParen, ")"})
	expected = append(expected, expectedResult{tokenError, "apparent function call missing ("})
	verifyLexerExprResults(t, input, expected)
}

func TestLexerExprNewline(t *testing.T) {
	input := `123
$foo`
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenInteger, "123"})
	expected = append(expected, expectedResult{tokenError, "newline, semicolon, and hash not allowed in expression"})
	verifyLexerExprResults(t, input, expected)
}

func TestLexerExprComment(t *testing.T) {
	input := "123 # foo"
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenInteger, "123"})
	expected = append(expected, expectedResult{tokenError, "newline, semicolon, and hash not allowed in expression"})
	verifyLexerExprResults(t, input, expected)
}

func TestLexerExprSemicolon(t *testing.T) {
	input := "123 ; foo"
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenInteger, "123"})
	expected = append(expected, expectedResult{tokenError, "newline, semicolon, and hash not allowed in expression"})
	verifyLexerExprResults(t, input, expected)
}

func TestLexerTokenConents(t *testing.T) {
	input := `$foo ${foo} [foo] "foo [bar] baz quux" {foo bar} foobar`
	expected := make([]expectedResult, 0)
	expected = append(expected, expectedResult{tokenVariable, "foo"})
	expected = append(expected, expectedResult{tokenVariable, "foo"})
	expected = append(expected, expectedResult{tokenCommand, "foo"})
	expected = append(expected, expectedResult{tokenQuote, "foo "})
	expected = append(expected, expectedResult{tokenCommand, "bar"})
	expected = append(expected, expectedResult{tokenQuote, " baz quux"})
	expected = append(expected, expectedResult{tokenBrace, "foo bar"})
	expected = append(expected, expectedResult{tokenString, "foobar"})
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
}
