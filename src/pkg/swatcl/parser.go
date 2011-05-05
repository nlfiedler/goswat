//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

// NewParser constructs a new Parser ready to parse the given Tcl text.
func NewParser(text string) *Parser {
	p := new(Parser)
	p.text = text
	p.len = len(text)
	p.token = tokenEOL
	return p
}

// GetTokenText returns the text of the current token.
func (p *Parser) GetTokenText() string {
	tlen := p.end - p.start + 1
	if tlen < 0 {
		tlen = 0
	}
	return p.text[p.start : p.start+tlen]
}

// parseSep expects the current position to be the start of a separator
// and advances until it finds the end of that separator.
func (p *Parser) parseSep() (parserState, *TclError) {
	p.start = p.p
	for p.len > 0 && (p.text[p.p] == ' ' || p.text[p.p] == '\t' ||
		p.text[p.p] == '\n' || p.text[p.p] == '\r') {
		p.p++
		p.len--
	}
	p.end = p.p - 1
	p.token = tokenSeparator
	return stateOK, nil
}

// parseEol expects the current position to be near the end of the line
// and advances until it finds the actual end of the line, passing over
// all whitespace and semicolons.
func (p *Parser) parseEol() (parserState, *TclError) {
	p.start = p.p
	for p.len > 0 && (p.text[p.p] == ' ' || p.text[p.p] == '\t' ||
		p.text[p.p] == '\n' || p.text[p.p] == '\r' || p.text[p.p] == ';') {
		p.p++
		p.len--
	}
	p.end = p.p - 1
	p.token = tokenEOL
	return stateOK, nil
}

// parseComment expects the current position to be the start of a
// comment and advances until it finds the end of the line.
func (p *Parser) parseComment() (parserState, *TclError) {
	for p.len > 0 && p.text[p.p] != '\n' && p.text[p.p] != '\r' {
		p.p++
		p.len--
	}
	return stateOK, nil
}

// parseCommand expects the current position to be an open square
// bracket ([) and advances to the end of the command (marked by a
// closing square bracket (]).
func (p *Parser) parseCommand() (parserState, *TclError) {
	// skip over the initial open bracket
	level := 1
	blevel := 0
	p.p++
	p.len--
	p.start = p.p
	for p.len > 0 {
		if p.text[p.p] == '[' && blevel == 0 {
			level++
		} else if p.text[p.p] == ']' && blevel == 0 {
			level--
			if level == 0 {
				break
			}
		} else if p.text[p.p] == '\\' {
			// pass over escaped characters
			p.p++
			p.len--
		} else if p.text[p.p] == '{' {
			blevel++
		} else if p.text[p.p] == '}' {
			if blevel != 0 {
				blevel--
			}
		}
		p.p++
		p.len--
	}
	p.end = p.p - 1
	p.token = tokenCommand
	if p.len > 0 && p.text[p.p] == ']' {
		p.p++
		p.len--
	}
	return stateOK, nil
}

// parseVariable expects the current position to be a dollar sign ($)
// and advances to the end of the variable reference. If the dollar sign
// is simply a bare character, it is treated as a string.
func (p *Parser) parseVariable() (parserState, *TclError) {
	// skip over the initial dollar sign ($)
	p.p++
	p.len--
	p.start = p.p
	braced := false
	if p.len > 0 && p.text[p.p] == '{' {
		// The variable is of the form ${foo}
		p.start++
		p.p++
		p.len--
		braced = true
	}
	for p.len > 0 && ((p.text[p.p] >= 'a' && p.text[p.p] <= 'z') ||
		(p.text[p.p] >= 'A' && p.text[p.p] <= 'Z') ||
		(p.text[p.p] >= '0' && p.text[p.p] <= '9') || p.text[p.p] == '_') {
		p.p++
		p.len--
	}
	if braced {
		// ensure there is a matching closing brace
		if p.len > 0 && p.text[p.p] == '}' {
			p.p++
			p.len--
		} else {
			return stateError, NewTclError(EBRACE, "missing closing brace")
		}
	}
	if p.start == p.p {
		// It's just a single char string '$'
		p.end = p.p - 1
		p.start = p.end
		p.token = tokenString
	} else if braced {
		p.end = p.p - 2
		p.token = tokenVariable
	} else {
		p.end = p.p - 1
		p.token = tokenVariable
	}
	return stateOK, nil
}

// parseBrace expects the current position to be an open curly brace ({)
// and advances to the subsequent closing brace, including any enclosed
// open and closing curly braces.
func (p *Parser) parseBrace() (parserState, *TclError) {
	// skip over the initial open brace
	level := 1
	p.p++
	p.len--
	p.start = p.p
	for {
		if p.len >= 2 && p.text[p.p] == '\\' {
			p.p++
			p.len--
		} else if p.len == 0 || p.text[p.p] == '}' {
			level--
			if level == 0 || p.len == 0 {
				p.end = p.p - 1
				if p.len > 0 {
					// Skip final closed brace
					p.p++
					p.len--
				}
				p.token = tokenString
				return stateOK, nil
			}
		} else if p.text[p.p] == '{' {
			level++
		}
		p.p++
		p.len--
	}
	panic("reached unreachable code")
}

// parseString processes the text at the current location as if it were
// (part of) a string, returning an appropriate state based on whether
// the string is brace enclosed or double quoted and contains nested
// substitutions.
func (p *Parser) parseString() (parserState, *TclError) {
	// determine the state of things
	newword := p.token == tokenSeparator || p.token == tokenEOL || p.token == tokenString
	if newword && p.text[p.p] == '{' {
		// braces are read without interpretation
		return p.parseBrace()
	} else if newword && p.text[p.p] == '"' {
		// quoted strings are subjected to substitutions
		p.insidequote = true
		p.p++
		p.len--
	}
	p.start = p.p
	for {
		if p.len == 0 {
			p.end = p.p - 1
			p.token = tokenString
			return stateOK, nil
		}
		switch p.text[p.p] {
		case '\\':
			if p.len >= 2 {
				// pass over escaped characters
				p.p++
				p.len--
			}
		case '$', '[':
			p.end = p.p - 1
			p.token = tokenEscape
			return stateOK, nil
		case ' ', '\t', '\n', '\r', ';':
			if !p.insidequote {
				p.end = p.p - 1
				p.token = tokenString
				return stateOK, nil
			}
		case '"':
			if p.insidequote {
				p.end = p.p - 1
				p.token = tokenString
				p.p++
				p.len--
				p.insidequote = false
				return stateOK, nil
			}
		}
		p.p++
		p.len--
	}
	panic("reached unreachable code")
}

// parseToken evaluates the token at the current position and parses
// that token appropriately, returning the parser in a state such that
// it indicates the type of the token, and the start/end points of the
// token.
func (p *Parser) parseToken() (parserState, *TclError) {
	for {
		if p.len == 0 {
			if p.token != tokenEOL && p.token != tokenEOF {
				p.token = tokenEOL
			} else {
				p.token = tokenEOF
			}
			return stateOK, nil
		}
		switch p.text[p.p] {
		case ' ', '\t', '\r':
			if p.insidequote {
				return p.parseString()
			}
			return p.parseSep()
		case '\n', ';':
			if p.insidequote {
				return p.parseString()
			}
			return p.parseEol()
		case '[':
			return p.parseCommand()
		case '$':
			return p.parseVariable()
		case '#':
			if p.token == tokenEOL {
				p.parseComment()
				continue
			}
			return p.parseString()
			// TODO: handle trailing backslash (it and newline are replaced by a space)
		default:
			return p.parseString()
		}
	}
	panic("reached unreachable code")
}

// parseExprToken is similar to parseToken and differs in that it
// expects to be parsing an expression, as taken by the expr command.
// Such expressions do not contain newlines, semicolons, or comments.
// Expressions may include function invocations and operators and their
// operands may be grouped within matching parentheses.
func (p *Parser) parseExprToken() (parserState, *TclError) {
	// Expressions do not have bare strings, so anything not
	// inside quotes, braces, or brackets must be an operator,
	// function call, or a number.
	for {
		if p.len == 0 {
			p.token = tokenEOF
			return stateOK, nil
		}
		switch p.text[p.p] {
		case ' ', '\t', '\r':
			if p.insidequote {
				return p.parseString()
			}
			return p.parseSep()
		case '\n', ';', '#':
			return stateError, NewTclError(EBADEXPR,
				"newline, semicolon, and hash not allowed in expression")
		case '[':
			return p.parseCommand()
		case '$':
			return p.parseVariable()
		case '"':
			return p.parseString()
		case '{':
			return p.parseBrace()
		case '(':
			p.start = p.p
			p.end = p.p
			p.p++
			p.len--
			p.token = tokenParen
			return stateOK, nil
		case '-', '+', '~', '!', '*', '/', '%', '<', '>', '=', '&', '^', '|', '?', ':':
			return p.parseOperator()
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return p.parseNumber()
		case ',':
			p.start = p.p
			p.end = p.p
			p.p++
			p.len--
			p.token = tokenComma
			return stateOK, nil
		default:
			// Must be a function call
			return p.parseFunction()
		}
	}
	panic("reached unreachable code")
}

// parseOperator expects the current position to be an operator and
// advances one or two characters depending on the width of the
// operator.
func (p *Parser) parseOperator() (parserState, *TclError) {
	p.start = p.p
	// skip over first operator character
	p.p++
	p.len--
	// check for two character operator (**, <<, >>, <=, >=, &&, ||)
	if p.len > 0 && (p.text[p.p] == '*' || p.text[p.p] == '>' ||
		p.text[p.p] == '<' || p.text[p.p] == '=' ||
		p.text[p.p] == '&' || p.text[p.p] == '|') {
		p.p++
		p.len--
	}
	p.end = p.p - 1
	p.token = tokenOperator
	return stateOK, nil
}

// parseNumber expects the current position to be the start of a numeric
// literal, and advances to the end of the literal.
func (p *Parser) parseNumber() (parserState, *TclError) {
	p.start = p.p
	float := false

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

	if p.len > 2 && p.text[p.p] == '0' && p.text[p.p+1] == 'x' {
		p.p += 2
		p.len -= 2
		// hexadecimal integer, scan as such
		for p.len > 0 && ((p.text[p.p] >= '0' && p.text[p.p] <= '9') ||
			(p.text[p.p] >= 'a' && p.text[p.p] <= 'f') ||
			(p.text[p.p] >= 'A' && p.text[p.p] <= 'F')) {
			p.p++
			p.len--
		}
	} else if p.len > 2 && p.text[p.p] == '0' && p.text[p.p+1] != '.' {
		// octal integer, scan as such
		for p.len > 0 && p.text[p.p] >= '0' && p.text[p.p] <= '7' {
			p.p++
			p.len--
		}

	} else {
		// it is either decimal integer or floating point
		sawexp := false
		sawsign := false
		for p.len > 0 {
			if p.text[p.p] == '.' {
				float = true
			} else if p.text[p.p] == 'e' || p.text[p.p] == 'E' {
				float = true
				sawexp = true
			} else if p.text[p.p] == '+' || p.text[p.p] == '-' {
				if !sawexp {
					// reached the end of the number
					break
				}
				sawsign = true
			} else if p.text[p.p] >= '0' && p.text[p.p] <= '9' {
				if sawsign || sawexp {
					// reset the flags now that we found a digit
					sawsign = false
					sawexp = false
				}
			} else {
				break
			}
			p.p++
			p.len--
		}
		if sawsign || sawexp {
			// if still set, these indicate a malformed expression
			return stateError, NewTclError(EBADEXPR, "malformed number")
		}
	}

	p.end = p.p - 1
	if float {
		p.token = tokenFloat
	} else {
		p.token = tokenInteger
	}
	return stateOK, nil
}

// parseFunction expects the current position to be the start of a
// function invocation, and advances to the opening parenthesis.
func (p *Parser) parseFunction() (parserState, *TclError) {
	p.start = p.p
	// Scan forward as long as we see an alphanumeric string,
	// stopping once we reach the open parenthesis.
	sawparen := false
	for p.len > 0 && ((p.text[p.p] >= '0' && p.text[p.p] <= '9') ||
		(p.text[p.p] >= 'a' && p.text[p.p] <= 'z') ||
		(p.text[p.p] >= 'A' && p.text[p.p] <= 'Z') ||
		p.text[p.p] == '(') {
		p.p++
		p.len--
		if p.text[p.p-1] == '(' {
			sawparen = true
			break
		}
	}
	if !sawparen {
		return stateError, NewTclError(EBADEXPR, "apparent function call missing ()")
	}
	p.end = p.p - 1
	p.token = tokenFunction
	return stateOK, nil
}
