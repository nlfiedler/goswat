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
	_             = iota
	stateOK       // evaluation successful
	stateError    // evaluation error, check returned error
	stateReturn   // 'return' command status
	stateBreak    // 'break' command status
	stateContinue // 'continue' command status
)

type parserToken int

const (
	_              = iota
	tokenEscape    // escape token
	tokenString    // string token
	tokenCommand   // command token
	tokenVariable  // variable token
	tokenSeparator // separator token
	tokenEOL       // end-of-line token
	tokenEOF       // end-of-file token
)

// Parser represents the internal state of the Tcl parser, including the
// text being parsed and the current token.
type Parser struct {
	text        string      // the text being parsed
	p           int         // current text position
	len         int         // remaining length to be parsd
	start       int         // start of current token
	end         int         // end of current token
	token       parserToken // token type (one of the token* constants)
	insidequote bool        // true if inside quotes
}

// callFrame is a frame within the call stack of the Tcl interpreter.
type callFrame struct {
	vars map[string]string
}

// commandFunc is a function that implements a built-in command. The
// argv parameter provides the incoming arguments, with the first entry
// being the name of the command being invoked. The data parameter is
// that which was passed to the RegisterCommand method of Interpreter.
// The function returns the parser state and the result of the command.
type commandFunc func(i *Interpreter, argv []string, data []string) (parserState, string)

// swatclCmd represents a built-in command.
type swatclCmd struct {
	function commandFunc // the command function
	privdata []string    // private data given at time of registration
}

// Interpreter contains the internal state of the Tcl interpreter,
// including register commands, the call frame, and result of the
// interpretation.
type Interpreter struct {
	level    int                  // level of nesting
	frames   []callFrame          // call stack frames
	commands map[string]swatclCmd // registered commands
	result   string               // result of evaluation
}

// Error constants
const (
	_         = iota
	EBRACE    // EBRACE indicates an unmatched curly brace ({)
	ECMDDEF   // indicates command is already defined
	EVARUNDEF // variable not defined
	ECMDUNDEF // command not defined
	ENOSTACK  // no call frames on the stack
	EILLARG   // interpreter given illegal method arguments
	EBADBOOL  // interpreter given a malformed boolean value
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
