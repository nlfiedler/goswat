//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"fmt"
	"io"
	"os"
)

// callFrame is a frame within the call stack of the Tcl interpreter.
type callFrame struct {
	vars map[string]string
}

// returnCode is used by commands to indicate their exit status.
type returnCode int

const (
	returnOk       returnCode = iota // return code TCL_OK
	returnError                      // return code TCL_ERROR
	returnReturn                     // return code TCL_RETURN
	returnBreak                      // return code TCL_BREAK
	returnContinue                   // return code TCL_CONTINUE
)

// commandFunc is a function that implements a built-in command. The
// argv parameter provides the incoming arguments, with the first entry
// being the name of the command being invoked. The data parameter is
// that which was passed to the RegisterCommand method of Interpreter.
// The function returns the result of the command and any error.
type commandFunc func(i Interpreter, argv []string, data []string) (string, returnCode, *TclError)

// swatclCmd represents a built-in command.
type swatclCmd struct {
	function commandFunc // the command function
	privdata []string    // private data given at time of registration
}

// Interpreter contains the internal state of the Tcl interpreter, including
// registered commands, and the call stack.
type Interpreter interface {
	io.Writer
	// GetVariable retrieves a variable from the current call frame.
	GetVariable(name string) (string, *TclError)
	// SetVariable sets or updates a variable in the current call frame.
	SetVariable(name, value string) *TclError
	// RegisterCommand adds the named command function to the interpreter so
	// it may be invoked at a later time.
	RegisterCommand(name string, f commandFunc, privdata []string) *TclError
	// InvokeCommand will call the named command, passing the given arguments.
	InvokeCommand(argv []string) (string, returnCode, *TclError)
	// Evaluate interprets the given Tcl text.
	Evaluate(tcl string) (string, returnCode, *TclError)
	// SetOutput changes the default output stream for commands like puts.
	// Passing a nil value will reset to the interpreter's default stdout.
	SetOutput(out io.Writer) *TclError
	// addFrame adds an empty call frame to the interpreter.
	addFrame()
	// popFrame removes the top-most frame from the stack.
	popFrame()
}

// interpreter is an implementation of the Interpreter interface.
type interpreter struct {
	level    int                  // level of nesting
	frames   []callFrame          // call stack frames
	commands map[string]swatclCmd // registered commands
	stdout   io.Writer            // where output generally goes
}

// NewInterpreter creates a new instance of Interpreter.
func NewInterpreter() Interpreter {
	i := new(interpreter)
	i.frames = make([]callFrame, 0)
	i.addFrame()
	i.commands = make(map[string]swatclCmd, 0)
	i.stdout = os.Stdout
	registerCoreCommands(i)
	return i
}

// addFrame adds an empty call frame to the interpreter.
func (i *interpreter) addFrame() {
	m := make(map[string]string)
	f := callFrame{m}
	i.frames = append(i.frames, f)
}

// popFrame removes the top-most frame from the stack.
func (i *interpreter) popFrame() {
	last := len(i.frames)
	i.frames = i.frames[:last-1]
}

// GetVariable retrieves a variable from the current call frame.
func (i *interpreter) GetVariable(name string) (string, *TclError) {
	last := len(i.frames)
	if last == 0 {
		err := fmt.Sprintf("Empty call stack, cannot get '%s'", name)
		return "", NewTclError(ENOSTACK, err)
	}
	frame := i.frames[last-1]
	v, ok := frame.vars[name]
	if !ok {
		err := fmt.Sprintf("Variable '%s' undefined", name)
		return "", NewTclError(EVARUNDEF, err)
	}
	return v, nil
}

// SetVariable sets or updates a variable in the current call frame.
func (i *interpreter) SetVariable(name, value string) *TclError {
	last := len(i.frames)
	if last == 0 {
		err := fmt.Sprintf("Empty call stack, cannot set '%s'", name)
		return NewTclError(ENOSTACK, err)
	}
	frame := i.frames[last-1]
	frame.vars[name] = value
	return nil
}

// RegisterCommand adds the named command function to the interpreter so
// it may be invoked at a later time.
func (i *interpreter) RegisterCommand(name string, f commandFunc, privdata []string) *TclError {
	_, ok := i.commands[name]
	if ok {
		err := fmt.Sprintf("Command '%s' already defined", name)
		return NewTclError(ECMDDEF, err)
	}
	cmd := swatclCmd{f, privdata}
	i.commands[name] = cmd
	return nil
}

// InvokeCommand will call the named command, passing the given arguments.
// The result of the command, as well as any error, are returned.
func (i *interpreter) InvokeCommand(argv []string) (string, returnCode, *TclError) {
	if len(argv) < 1 {
		err := "InvokeCommand called without arguments"
		return "", returnError, NewTclError(EILLARG, err)
	}
	name := argv[0]
	c, ok := i.commands[name]
	if !ok {
		err := fmt.Sprintf("No such command '%s'", name)
		return "", returnError, NewTclError(ECMDUNDEF, err)
	}
	return c.function(i, argv, c.privdata)
}

// Evaluate interprets the given Tcl text.
func (i *interpreter) Evaluate(tcl string) (string, returnCode, *TclError) {
	// command and arguments to be invoked
	argv := make([]string, 0)
	c := lex("Evaluate", tcl)

	// TODO: handle escaped newline at end of string (inside both " and {, converts to space)
	inquotes := false
	result := ""
	var code returnCode
	var err *TclError
	for {
		token, ok := <-c
		if !ok {
			return "", returnError, NewTclError(ELEXER, "unexpected end of lexer stream")
		}
		result := token.contents()
		switch token.typ {
		case tokenError:
			return "", returnError, NewTclError(ELEXER, token.val)
		case tokenEOL, tokenEOF:
			if len(argv) > 0 {
				// Parsing complete, invoke the command.
				result, code, err = i.InvokeCommand(argv)
				if err != nil {
					return result, code, err
				} else if token.typ == tokenEOF {
					return result, code, nil
				}
				argv = make([]string, 0)
			}

		case tokenVariable:
			// Get variable value
			result, err = i.GetVariable(result)
			if err != nil {
				return "", returnError, err
			}

		case tokenCommand:
			// Evaluate command invocation
			result, code, err = i.Evaluate(result)
			if err != nil {
				return "", code, err
			}

		case tokenQuote:
			qb, qe := token.quotes()
			if qb != qe {
				inquotes = !inquotes
			}
		}

		// We have a new token, append to the previous or as new arg?
		if inquotes {
			last := len(argv) - 1
			argv[last] = argv[last] + result
		} else {
			argv = append(argv, result)
		}
	}
	return result, returnOk, nil
}

// Write writes the given bytes to the standard output for this interpreter.
func (i *interpreter) Write(p []byte) (n int, err error) {
	return i.stdout.Write(p)
}

// SetOutput makes the given writer be the output for this interpreter.
func (i *interpreter) SetOutput(out io.Writer) *TclError {
	if _, ok := out.(io.Writer); ok {
		i.stdout = out
	} else {
		// allow passing nil to reset the output
		i.stdout = os.Stdout
	}
	return nil
}

// registerCoreCommands registers the built-in commands provided by this
// package so that they may be used by other Tcl scripts.
func registerCoreCommands(i Interpreter) {
	i.RegisterCommand("expr", commandExpr, nil)
	i.RegisterCommand("if", commandIf, nil)
	i.RegisterCommand("puts", commandPuts, nil)
	i.RegisterCommand("set", commandSet, nil)
}

// arityError is a convenience method for commands to report an error
// with the number of arguments given to the command.
func arityError(name string) *TclError {
	return NewTclError(ECOMMAND, "Wrong number of arguments for "+name)
}
