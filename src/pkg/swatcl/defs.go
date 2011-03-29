//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

// The swatcl package implements a rudimentary Tcl interpreter for the
// purpose of writing debugger commands.
package swatcl

import (
	"os"
)

type parserState int

const (
	_ = iota
	// evaluation successful
	stateOK
	// evaluation error, check returned error
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
	tokenEscape
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

// Parser represents the internal state of the Tcl parser, including the
// text being parsed and the current token.
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

// callFrame is a frame within the call stack of the Tcl interpreter.
type callFrame struct {
	vars map[string]string
}

// commandFunc is a function that implements a built-in command.
type commandFunc func(context *Interpreter, argv []string, data []byte) parserState

// swatclCmd represents a built-in command.
type swatclCmd struct {
	function commandFunc
	privdata []byte
}

// Interpreter contains the internal state of the Tcl interpreter,
// including register commands, the call frame, and result of the
// interpretation.
type Interpreter struct {
	// Level of nesting
	level    int
	frames   []callFrame
	commands map[string]swatclCmd
	result   string
}

// Error constants
const (
	_         = iota
	EBRACE    // EBRACE indicates an unmatched curly brace ({)
	ECMDDEF   // indicates command is already defined
	EVARUNDEF // variable not defined
	ECMDUNDEF // command not defined
	ENOSTACK  // no call frames on the stack
)

// TclError is used to provide information on the type of error that
// occurred while parsing or evaluating the Tcl script.
type TclError struct {
	Errno os.Errno
	Error os.Error
}

// NewTclError creates a new TclError based on the given values.
func NewTclError(err int, msg string) *TclError {
	return &TclError{os.Errno(err), os.ErrorString(msg)}
}

// String returns the string representation of the error.
func (e *TclError) String() string {
	return e.Error.String()
}
