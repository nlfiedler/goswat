//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"syscall"
)

// Error constants
const (
	_         = iota
	EOK       // no error
	EBRACE    // found unmatched curly brace ({)
	ECMDDEF   // command is already defined
	EVARUNDEF // variable not defined
	ECMDUNDEF // command not defined
	ENOSTACK  // no call frames on the stack
	EILLARG   // interpreter given illegal method arguments
	EBADBOOL  // interpreter given a malformed boolean value
	EBADEXPR  // invalid characters in expression
	EOPERAND  // missing or malformed operand
	EOPERATOR // invalid or unknown operator
	EBADSTATE // expression evaluator in a bad state
	EINVALNUM // invalid numeric expression
	ENUMRANGE // numeric value out of supported range
	ELEXER    // lexer tokenization failed
	ECOMMAND  // error related to commands
)

// TODO: read http://golang.org/doc/go_faq.html#nil_error and change all this
// TclError is used to provide information on the type of error that
// occurred while parsing or evaluating the Tcl script. It implements
// the error interface.
type TclError struct {
	Errno   syscall.Errno
	Message string
}

// NewTclError creates a new TclError based on the given values.
func NewTclError(err int, msg string) *TclError {
	return &TclError{syscall.Errno(err), msg}
}

// String returns the string representation of the error.
func (e *TclError) String() string {
	return e.Message
}

// Error returns the string representation of the error.
func (e *TclError) Error() string {
	return e.Message
}
