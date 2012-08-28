//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

//
// This lexer is fashioned after that which was presented by Rob Pike in the
// "Lexical Scanning in Go" talk (http://cuddle.googlecode.com/hg/talk/lex.html).
//

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// tokenType is the type of a lexer token (e.g. string, number).
type tokenType int

// eof marks the end of the input text.
const eof = unicode.UpperLower

// token types
const (
	_             tokenType = iota // error occurred
	tokenError                     // error occurred
	tokenString                    // string token
	tokenQuote                     // quoted string token
	tokenBrace                     // uninterpreted string token
	tokenCommand                   // command token
	tokenVariable                  // variable token
	tokenFunction                  // expression function call
	tokenOperator                  // expression operator
	tokenInteger                   // integer literal
	tokenFloat                     // floating point literal
	tokenComma                     // comma (argument separator)
	tokenParen                     // open/close parenthesis
	tokenEOL                       // end-of-line token
	tokenEOF                       // end-of-file token
)

// token represents a token returned from the scanner.
type token struct {
	typ tokenType // Type, such as tokenNumber.
	val string    // Value, such as "23.2".
}

// String returns the string representation of the lexer token.
func (t *token) String() string {
	switch t.typ {
	case tokenEOF:
		return "EOF"
	case tokenError:
		return t.val
	}
	if len(t.val) > 10 {
		return fmt.Sprintf("%.10q...", t.val)
	}
	return fmt.Sprintf("%q", t.val)
}

// quotes indicates whether the token value starts and ends with double
// quotes ("). The first boolean return value is true if the token value
// starts with ", while the second return value is true if the token
// value ends with ". If the token value is a single character, the
// second return value is always false. If the token is not a quoted
// token, the return values will be false, false.
func (t *token) quotes() (bool, bool) {
	if t.typ == tokenQuote && len(t.val) > 0 {
		l := len(t.val)
		if l == 1 {
			return t.val[0] == '"', false
		} else {
			return t.val[0] == '"', t.val[l-1] == '"'
		}
	} else {
		return false, false
	}
	panic("unreachable code")
}

// contents returns the unique portion of the token text, minus any
// markers such as braces or brackets. For tokenEOF this will return
// nil.
func (t *token) contents() string {
	switch t.typ {
	case tokenBrace, tokenCommand:
		return t.val[1 : len(t.val)-1]
	case tokenQuote:
		qb, qe := t.quotes()
		l := len(t.val)
		if qb {
			if qe {
				return t.val[1 : l-1]
			}
			return t.val[1:]
		} else if qe {
			return t.val[:l-1]
		}
		return t.val
	case tokenVariable:
		if t.val[1] == '{' {
			l := len(t.val)
			return t.val[2 : l-1]
		}
		return t.val[1:]
	default:
		return t.val
	}
	panic("unreachable code")
}

// lexer holds the state of the scanner.
type lexer struct {
	name   string     // used only for error reports.
	input  string     // the string being scanned.
	start  int        // start position of this token.
	pos    int        // current position in the input.
	width  int        // width of last rune read from input.
	state  stateFn    // starting state, others states should return here
	tokens chan token // channel of scanned tokens.
}

// stateFn represents the state of the scanner as a function that
// returns the next state.
type stateFn func(*lexer) stateFn

// lex initializes the lexer to lex the given Tcl command text,
// returning the channel from which tokens are received.
func lex(name, input string) chan token {
	l := &lexer{
		name:   name,
		input:  input,
		tokens: make(chan token),
		state:  lexStart,
	}
	go l.run() // Concurrently run state machine.
	return l.tokens
}

// lexExpr initializes the lexer to lex the given Tcl expression,
// returning the channel from which tokens are received.
func lexExpr(name, input string) chan token {
	l := &lexer{
		name:   name,
		input:  input,
		tokens: make(chan token),
		state:  lexExprStart,
	}
	go l.run() // Concurrently run state machine.
	return l.tokens
}

// run lexes the input by executing state functions until the state is
// nil, which marks the end of the input.
func (l *lexer) run() {
	for state := l.state; state != nil; {
		state = state(l)
	}
	close(l.tokens) // No more tokens will be delivered.
}

// emit passes a token back to the client via the channel.
func (l *lexer) emit(t tokenType) {
	l.tokens <- token{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune. Can be called only once per call to next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing back
// a nil pointer that will be the next state.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{
		tokenError,
		fmt.Sprintf(format, args...),
	}
	return l.state
}

// lexStart reads the next token from the input and determines
// what to do with that token, returning the appropriate state
// function.
func lexStart(l *lexer) stateFn {
	r := l.next()
	switch r {
	case eof:
		l.emit(tokenEOF)
		return nil
	case ' ', '\t', '\r':
		return lexSeparator
	case '\n', ';':
		return lexEol
	case '[':
		return lexCommand
	case '$':
		return lexVariable
	case '#':
		return lexComment
	case '{':
		return lexBrace
	case '"':
		return lexQuotes
	default:
		return lexString
	}
	panic("unreachable code")
}

// lexQuotes expects the current character to be a double-quote and
// scans the input to find the end of the quoted string, possibly
// emitting string tokens as well as variable and command tokens.
func lexQuotes(l *lexer) stateFn {
	for {
		r := l.next()
		switch r {
		case eof:
			return l.errorf("unclosed quoted string: %q", l.input[l.start:l.pos])
		case '\\':
			// pass over escaped characters
			l.next()
		case '$':
			l.backup()
			l.emit(tokenQuote)
			l.next()
			lexVariable(l)
		case '[':
			l.backup()
			l.emit(tokenQuote)
			l.next()
			lexCommand(l)
		case '"':
			// reached the end of the string
			l.emit(tokenQuote)
			return l.state
		}
	}
	panic("unreachable code")
}

// lexSeparator expects the current position to be the start of a
// separator and advances until it finds the end of that separator.
// No token will be emitted since separators are ignored.
func lexSeparator(l *lexer) stateFn {
	l.acceptRun(" \t\n\r")
	l.ignore()
	return l.state
}

// lexEol expects the current position to be near the end of the line
// and advances until it finds the actual end of the line, passing over
// all whitespace and semicolons.
func lexEol(l *lexer) stateFn {
	l.acceptRun(" \t\n\r;")
	l.emit(tokenEOL)
	return l.state
}

// lexComment expects the current position to be the start of a
// comment and advances until it finds the end of the line/file.
// No token will be emitted since comments are ignored.
func lexComment(l *lexer) stateFn {
	for {
		r := l.next()
		switch r {
		case eof:
			l.ignore()
			return l.state
		case '\n', '\r':
			l.backup()
			l.ignore()
			return lexEol
		}
	}
	panic("unreachable code")
}

// lexBrace expects the current position to be an open curly brace ({)
// and advances to the subsequent closing brace, including any enclosed
// open and closing curly braces. Any variable or command references
// found within are ignored and treated as text.
func lexBrace(l *lexer) stateFn {
	level := 1 // open brace count
	for {
		r := l.next()
		switch r {
		case eof:
			return l.errorf("unclosed left brace: %q", l.input[l.start:l.pos])
		case '\\':
			// pass over escaped characters
			l.next()
		case '}':
			level--
			if level == 0 {
				l.emit(tokenBrace)
				return l.state
			}
		case '{':
			level++
		}
	}
	panic("unreachable code")
}

// lexString processes the text at the current location as if it were a
// string that is not enclosed within quotes. If any non-string
// characters are encountered, the string token is emitted and control
// returns to the starting state.
func lexString(l *lexer) stateFn {
	for {
		r := l.next()
		switch r {
		case eof:
			l.emit(tokenString)
			return l.state
		case '\\':
			// pass over escaped characters
			l.next()
		case '{', '$', '[', ' ', '\t', '\n', '\r', ';':
			// reached the end of the string
			l.backup()
			l.emit(tokenString)
			return l.state
		}
	}
	panic("unreachable code")
}

// lexCommand expects the current position to be an open square
// bracket ([) and advances to the end of the command (marked by a
// closing square bracket (]).
func lexCommand(l *lexer) stateFn {
	level := 1  // open command count
	blevel := 0 // open brace count
	for {
		r := l.next()
		if r == eof {
			return l.errorf("unclosed command: %q", l.input[l.start:l.pos])
		} else if r == '[' && blevel == 0 {
			level++
		} else if r == ']' && blevel == 0 {
			level--
			if level == 0 {
				break
			}
		} else if r == '\\' {
			// pass over escaped characters
			l.next()
		} else if r == '{' {
			blevel++
		} else if r == '}' && blevel != 0 {
			blevel--
		}
	}
	l.emit(tokenCommand)
	return l.state
}

// lexVariable expects the current position to be a dollar sign ($) and
// advances to the end of the variable reference. If the dollar sign is
// not followed by curly brace ({) or letters and numbers, it is treated
// as a string.
func lexVariable(l *lexer) stateFn {
	braced := false
	r := l.next()
	if r == '{' {
		// variable is of the form ${foo}
		r = l.next()
		braced = true
	}
	// scan to the end of the variable reference
	for (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
		r = l.next()
	}
	if braced {
		if r != '}' {
			return l.errorf("unclosed variable reference: %q", l.input[l.start:l.pos])
		}
	} else {
		l.backup()
	}
	if l.start == l.pos-1 {
		// it's not a variable reference
		return lexString
	}
	l.emit(tokenVariable)
	return l.state
}

// lexExprStart is similar to lexStart and differs in that it expects to
// be lexing an expression, as taken by the expr command. Such
// expressions do not contain newlines, semicolons, or comments.
// Expressions may include function invocations and operators and their
// operands may be grouped within matching parentheses.
func lexExprStart(l *lexer) stateFn {
	// Expressions do not have bare strings, so anything not
	// inside quotes, braces, or brackets must be an operator,
	// function call, or a number.
	r := l.next()
	switch r {
	case eof:
		l.emit(tokenEOF)
		return nil
	case ' ', '\t', '\r':
		return lexSeparator
	case '\n', ';', '#':
		return l.errorf("newline, semicolon, and hash not allowed in expression")
	case '[':
		return lexCommand
	case '$':
		return lexVariable
	case '{':
		return lexBrace
	case '"':
		return lexQuotes
	case '(', ')':
		l.emit(tokenParen)
		return l.state
	case '-', '+', '~', '!', '*', '/', '%', '<', '>', '=', '&', '^', '|', '?', ':':
		return lexOperator
	case '.', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		// let lexNumber sort out what type of number it is
		l.backup()
		return lexNumber
	case ',':
		l.emit(tokenComma)
		return l.state
	default:
		// must be a function call (let lexFunction assert that)
		l.backup()
		return lexFunction
	}
	panic("unreachable code")
}

// lexNumber expects the current position to be the start of a numeric
// literal, and advances to the end of the literal.
func lexNumber(l *lexer) stateFn {

	//
	// Can expect integer in octal, decimal, or hexadecimal, as well
	// as floating point, with optional exponent. Any leading sign
	// is handled elsewhere as a unary minus operator.
	//
	// 1
	// 2.1
	// 3.
	// 6E4
	// 7.91e+16
	// .000001
	// 0366 (octal)
	// 0x7b5 (hexadecimal)
	//

	float := false
	if l.accept("0") && l.accept("xX") {
		// hexadecimal
		l.acceptRun("0123456789abcdefABCDEF")
	} else if l.accept("0") && !l.accept(".") {
		// octal
		l.acceptRun("01234567")
	} else {
		// decimal, possibly floating point
		digits := "0123456789"
		l.acceptRun(digits)
		if l.accept(".") {
			float = true
			l.acceptRun(digits)
		}
		if l.accept("eE") {
			float = true
			l.accept("+-")
			l.acceptRun(digits)
		}
	}

	// Next thing must _not_ be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		return l.errorf("malformed number: %q", l.input[l.start:l.pos])
	}
	if float {
		l.emit(tokenFloat)
	} else {
		l.emit(tokenInteger)
	}
	return l.state
}

// lexOperator expects the current position to be an operator and
// advances one or two characters depending on the width of the
// operator.
func lexOperator(l *lexer) stateFn {
	// check for two character operator (**, <<, >>, <=, >=, &&, ||)
	l.accept("*<>=&|")
	l.emit(tokenOperator)
	return l.state
}

// lexFunction expects the next position to be the start of a function
// invocation, and advances to the opening parenthesis. Both a function
// and open parenthesis token will be emitted to the channel.
func lexFunction(l *lexer) stateFn {
	r := l.next()
	// scan forward as long as we see an alphanumeric string
	for (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
		r = l.next()
	}
	// next character must be an open parenthesis
	if r != '(' {
		return l.errorf("apparent function call missing (: %q", l.input[l.start:l.pos])
	}
	l.backup()
	l.emit(tokenFunction)
	l.next()
	l.emit(tokenParen)
	return l.state
}
