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
	return p.text[p.start:p.start + tlen]
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
	// skip over the initial open bracket ([)
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
	// skip over the initial open bracket
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
				p.token = tokenString;
				return stateOK, nil
			}
		} else if (p.text[p.p] == '{') {
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
			p.token = tokenEscape
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
				p.token = tokenEscape
				return stateOK, nil
			}
		case '"':
			if p.insidequote {
				p.end = p.p - 1
				p.token = tokenEscape
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
// it indicates the type of the token, the the start/end points of the
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
		default:
			return p.parseString()
		}
	}
	panic("reached unreachable code")
}
