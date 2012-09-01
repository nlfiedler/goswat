//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"errors"
	"fmt"
)

//
// Internal errors are those which are raised within the interpreter to signal
// an illegal action or invalid state. These generally should not be surfaced
// to the user as-is.
//

var BooleanInvalidSyntax = errors.New("swatcl: boolean syntax invalid")
var CallStackEmtpy = errors.New("swatcl: call stack is empty")
var CommandAlreadyDefined = errors.New("swatcl: command already defined")
var NumberSyntaxInvalid = errors.New("swatcl: number syntax invalid")
var NumberOutOfRange = errors.New("swatcl: number value of out range")
var VariableUndefined = errors.New("swatcl: variable undefined")

//
// The Tcl interpreter results and errors defined below are intended for human
// consumption, as they indicate something that is wrong with the Tcl script
// that is being evaluated.
//

// ErrorCode indicates the error status of the Tcl command, with EOK
// representing no error (i.e. command was successful).
type ErrorCode int

const (
	EOK       ErrorCode = iota // no error
	EARGUMENT                  // e.g. illegal, missing
	EBADSTATE                  // expression evaluator in a bad state
	ECOMMAND                   // e.g. undefined, unsupported, unknown
	ESYNTAX                    // e.g. invalid number syntax
	EVARIABLE                  // e.g. undefined
)

// returnCode is used by commands to indicate their exit status.
type returnCode int

const (
	returnOk       returnCode = iota // return code TCL_OK
	returnError                      // return code TCL_ERROR
	returnReturn                     // return code TCL_RETURN
	returnBreak                      // return code TCL_BREAK
	returnContinue                   // return code TCL_CONTINUE
)

// TclResult is returned from all Tcl commands, operators, and functions
// (as in commands the mathfunc namespace that implement math functions).
type TclResult interface {
	error
	// ErrorCode returns the error code associated with this result.
	ErrorCode() ErrorCode
	// ErrorMessage returns the error message associated with this result.
	ErrorMessage() string
	// Ok indicates if the result is a non-error, indicating that
	// the result is suitable for consumption.
	Ok() bool
	// Result returns the result of the evaluation.
	Result() string
	// ReturnCode returns the resulting return code (e.g. TCL_OK).
	// The values are the same as for the Tcl return codes.
	ReturnCode() returnCode
	// String returns a human readable error message.
	String() string
}

// tclResult implements the TclResult interface.
type tclResult struct {
	ecode  ErrorCode  // ecode indicates the type of the error, if any
	errmsg string     // errmsg is the error message, if any
	rcode  returnCode // rcode is the Tcl return code (e.g. TCL_OK, TCL_RETURN, etc)
	result string     // result is the result of the Tcl command
}

// newTclResultOk is a convenience function that constructs a TclResult in
// which everything is OK, there is no error message, and the result is as
// given.
func newTclResultOk(result string) TclResult {
	return newTclResult(EOK, "", returnOk, result)
}

// newTclResultError is a convenience function that constructs a TclResult in
// which the given error occurred and there is no meaningful result. The
// returnCode will be returnError.
func newTclResultError(code ErrorCode, detail string) TclResult {
	return newTclResult(code, detail, returnError, "")
}

// newTclResultError is a convenience function that constructs a TclResult in
// which the given error occurred and there is no meaningful result. The
// returnCode will be returnError.
func newTclResultErrorf(code ErrorCode, form string, args ...interface{}) TclResult {
	detail := fmt.Sprintf(form, args...)
	return newTclResult(code, detail, returnError, "")
}

// newTclResult constructs a TclResult instance based on the arguments.
func newTclResult(code ErrorCode, detail string, ret returnCode, result string) TclResult {
	return &tclResult{code, detail, ret, result}
}

// Returns the error portion of this result in string form.
func (r *tclResult) Error() string {
	return r.String()
}

// Error returns the error code, or EBADSTATE if undefined.
func (r *tclResult) ErrorCode() ErrorCode {
	if r != nil {
		return r.ecode
	}
	return EBADSTATE
}

// ErrorMessage returns the error message, if any.
func (r *tclResult) ErrorMessage() string {
	if r != nil {
		return r.errmsg
	}
	return ""
}

// Ok returns true if the error code is EOK, false otherwise.
func (r *tclResult) Ok() bool {
	if r != nil {
		return r.ecode == EOK
	}
	return false
}

// Result returns the result of the Tcl command, or "<nil>" if undefined.
func (r *tclResult) Result() string {
	if r != nil {
		return r.result
	}
	return "<nil>"
}

// ReturnCode returns the returnCode of the Tcl command, or returnError if
// undefined.
func (r *tclResult) ReturnCode() returnCode {
	if r != nil {
		return r.rcode
	}
	return returnError
}

// String returns a human readable for of the result.
func (r *tclResult) String() string {
	if r != nil {
		if r.Ok() {
			return r.Result()
		} else {
			// Would be nice to print code as text
			return fmt.Sprintf("ERR-%04d: %s", int(r.ecode), r.errmsg)
		}
	}
	return ""
}
