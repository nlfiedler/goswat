//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

type parserState int
const (
	_ = iota
	// evaluation successful
	stateOK parserState = iota
	// evaluation error
	stateError
	// 'return' command status
	stateReturn
	// 'break' command status
	stateBreak
	// 'continue' command status
	stateContinue
)

type parserToken int
const (
	_ = iota
	// escape token
	tokenEscape parserToken = iota
	// string token
	tokenString
	// command token
	tokenCommand
	// variable token
	tokenVariable
	// separator token
	tokenSeparator
	// end-of-line token
	tokenEOL
	// end-of-file token
	tokenEOF
)

type Parser struct {
	// the text being parsed
	text string
	// current text position
	p int
	// remaining length to be parsd
	len int
	// start of current token
	start int
	// end of current token
	end int
	// token type (one of the token* constants)
	token parserToken
	// true if inside quotes
	insidequote bool
}

type callFrame struct {
	vars map[string]string
}

type commandFunc func(context Interpreter) (int) //, int argc, char **argv, void *privdata);

type swatclCmd struct {
	name string
	function commandFunc
	privdata []byte
}

type Interpreter struct {
	// Level of nesting
	level int
	frames []callFrame
	commands []swatclCmd
	result string
}
