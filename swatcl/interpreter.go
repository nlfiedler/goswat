//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"io"
	"os"
	"strings"
)

// callFrame is a frame within the call stack of the Tcl interpreter.
type callFrame struct {
	vars map[string]string
}

// commandFunc is a function that implements a built-in command. The argv
// parameter provides the incoming arguments, with the first entry being the
// name of the command being invoked. The data parameter is that which was
// passed to the RegisterCommand method of Interpreter. The function returns
// the result of the command.
type commandFunc func(i Interpreter, argv []string, data []string) TclResult

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
	GetVariable(name string) (string, error)
	// SetVariable sets or updates a variable in the current call frame.
	SetVariable(name, value string) error
	// RegisterCommand adds the named command function to the interpreter so
	// it may be invoked at a later time.
	RegisterCommand(name string, f commandFunc, privdata []string) error
	// InvokeCommand will call the named command, passing the given arguments.
	InvokeCommand(argv []string) TclResult
	// Evaluate interprets the given Tcl text.
	Evaluate(tcl string) TclResult
	// SetOutput changes the default output stream for commands like puts.
	// Passing a nil value will reset to the interpreter's default stdout.
	SetOutput(out io.Writer) error
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
func (i *interpreter) GetVariable(name string) (string, error) {
	last := len(i.frames)
	if last == 0 {
		return "", CallStackEmtpy
	}
	frame := i.frames[last-1]
	v, ok := frame.vars[name]
	if !ok {
		return "", VariableUndefined
	}
	return v, nil
}

// SetVariable sets or updates a variable in the current call frame.
func (i *interpreter) SetVariable(name, value string) error {
	last := len(i.frames)
	if last == 0 {
		return CallStackEmtpy
	}
	frame := i.frames[last-1]
	frame.vars[name] = value
	return nil
}

// RegisterCommand adds the named command function to the interpreter so
// it may be invoked at a later time.
func (i *interpreter) RegisterCommand(name string, f commandFunc, privdata []string) error {
	_, ok := i.commands[name]
	if ok {
		return CommandAlreadyDefined
	}
	cmd := swatclCmd{f, privdata}
	i.commands[name] = cmd
	return nil
}

// InvokeCommand will call the named command, passing the given arguments.
// The result of the command, as well as any error, are returned.
func (i *interpreter) InvokeCommand(argv []string) TclResult {
	if len(argv) > 0 {
		name := argv[0]
		c, ok := i.commands[name]
		if !ok {
			return newTclResultErrorf(ECOMMAND, "No such command '%s'", name)
		}
		return c.function(i, argv, c.privdata)
	}
	return newTclResultOk("")
}

// Evaluate interprets the given Tcl text.
func (i *interpreter) Evaluate(tcl string) TclResult {
	// command and arguments to be invoked
	argv := make([]string, 0)
	c := lex("Evaluate", tcl)
	defer drainLexer(c)

	// TODO: handle escaped newline at end of string (for both "" and {}, converts to space)
	inquotes := false
	var result TclResult
	var err error
	for {
		token, ok := <-c
		if !ok {
			break
		}
		text := token.contents()
		switch token.typ {
		case tokenError:
			return newTclResultError(ESYNTAX, token.val)
		case tokenEOL, tokenEOF:
			if len(argv) > 0 {
				// Parsing complete, invoke the command.
				result = i.InvokeCommand(argv)
				if !result.Ok() || !result.ReturnOk() || token.typ == tokenEOF {
					return result
				}
				text = ""
				argv = make([]string, 0)
			}

		case tokenVariable:
			// Get variable value
			text, err = i.GetVariable(text)
			if err != nil {
				return newTclResultError(EVARIABLE, err.Error())
			}

		case tokenCommand:
			// Evaluate command invocation
			result = i.Evaluate(text)
			if !result.Ok() || !result.ReturnOk() {
				return result
			}
			text = result.Result()

		case tokenQuote:
			qb, qe := token.quotes()
			if qb != qe {
				inquotes = !inquotes
			}
		}

		text = strings.TrimSpace(text)
		if len(text) > 0 {
			// We have a new token, append to the previous or as new arg?
			if inquotes {
				last := len(argv) - 1
				argv[last] = argv[last] + text
			} else {
				argv = append(argv, text)
			}
		}
	}
	return result
}

// Write writes the given bytes to the standard output for this interpreter.
func (i *interpreter) Write(p []byte) (n int, err error) {
	return i.stdout.Write(p)
}

// SetOutput makes the given writer be the output for this interpreter.
func (i *interpreter) SetOutput(out io.Writer) error {
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
	i.RegisterCommand("break", commandBreak, nil)
	i.RegisterCommand("continue", commandContinue, nil)
	i.RegisterCommand("expr", commandExpr, nil)
	i.RegisterCommand("if", commandIf, nil)
	i.RegisterCommand("puts", commandPuts, nil)
	i.RegisterCommand("return", commandReturn, nil)
	i.RegisterCommand("set", commandSet, nil)
	i.RegisterCommand("while", commandWhile, nil)
}
