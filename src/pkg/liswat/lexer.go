//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package liswat

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
	tokenVariable                  // variable token
	tokenInteger                   // integer literal
	tokenFloat                     // floating point literal
	tokenParen                     // open/close parenthesis
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
	if t.typ == tokenString && len(t.val) > 0 {
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
	case tokenString:
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
	}
	go l.run() // Concurrently run state machine.
	return l.tokens
}

// run lexes the input by executing state functions until the state is
// nil, which marks the end of the input.
func (l *lexer) run() {
	for state := lexStart; state != nil; {
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
	return lexStart
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
	case '(', ')':
		l.emit(tokenParen)
		return lexStart
	case ' ', '\t', '\r', '\n':
		return lexSeparator
	case '.', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		// let lexNumber sort out what type of number it is
		l.backup()
		return lexNumber
	case ';':
		return lexComment
	case '"':
		return lexQuotes
	default:
		return lexVariable
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
		case '"':
			// reached the end of the string
			l.emit(tokenString)
			return lexStart
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
	return lexStart
}

// lexComment expects the current position to be the start of a
// comment and advances until it finds the end of the line/file.
// No token will be emitted since comments are ignored.
func lexComment(l *lexer) stateFn {
	for {
		r := l.next()
		switch r {
		case eof, '\n', '\r':
			l.ignore()
			return lexStart
		}
	}
	panic("unreachable code")
}

// lexVariable processes the text at the current location as if it were
// a variable reference.
func lexVariable(l *lexer) stateFn {
	for {
		r := l.next()
		switch r {
		case eof:
			return l.errorf("unexpectedly reached end at %q", l.input[l.start:l.pos])
		case '\\':
			// pass over escaped characters
			l.next()
		case '(', ')', ' ', '\t', '\n', '\r':
			// reached the end of the variable
			l.backup()
			l.emit(tokenVariable)
			return lexStart
		}
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
	return lexStart
}

// isAlphaNumeric indicates if the given rune is a letter or number.
func isAlphaNumeric(r rune) bool {
	return unicode.IsDigit(r) || unicode.IsLetter(r)
}
