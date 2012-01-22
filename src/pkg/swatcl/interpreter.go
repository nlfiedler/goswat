//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"fmt"
)

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

// NewInterpreter creates a new instance of Interpreter.
func NewInterpreter() *Interpreter {
	i := new(Interpreter)
	i.frames = make([]callFrame, 0)
	i.addFrame()
	i.commands = make(map[string]swatclCmd, 0)
	i.registerCoreCommands()
	return i
}

// addFrame adds an empty call frame to the interpreter.
func (i *Interpreter) addFrame() {
	m := make(map[string]string)
	f := callFrame{m}
	i.frames = append(i.frames, f)
}

// popFrame removes the top-most frame from the stack.
func (i *Interpreter) popFrame() {
	last := len(i.frames)
	i.frames = i.frames[:last-1]
}

// GetVariable retrieves a variable from the current call frame.
func (i *Interpreter) GetVariable(name string) (string, *TclError) {
	last := len(i.frames)
	if last == 0 {
		i.result = fmt.Sprintf("Empty call stack, cannot get '%s'", name)
		return "", NewTclError(ENOSTACK, i.result)
	}
	frame := i.frames[last-1]
	v, ok := frame.vars[name]
	if !ok {
		i.result = fmt.Sprintf("Variable '%s' undefined", name)
		return "", NewTclError(EVARUNDEF, i.result)
	}
	return v, nil
}

// SetVariable sets or updates a variable in the current call frame.
func (i *Interpreter) SetVariable(name, value string) *TclError {
	last := len(i.frames)
	if last == 0 {
		i.result = fmt.Sprintf("Empty call stack, cannot set '%s'", name)
		return NewTclError(ENOSTACK, i.result)
	}
	frame := i.frames[last-1]
	frame.vars[name] = value
	return nil
}

// RegisterCommand adds the named command function to the interpreter so
// it may be invoked at a later time.
func (i *Interpreter) RegisterCommand(name string, f commandFunc, privdata []string) (parserState, *TclError) {
	_, ok := i.commands[name]
	if ok {
		i.result = fmt.Sprintf("Command '%s' already defined", name)
		return stateError, NewTclError(ECMDDEF, i.result)
	}
	cmd := swatclCmd{f, privdata}
	i.commands[name] = cmd
	return stateOK, nil
}

// InvokeCommand will call the named command, passing the given arguments.
func (i *Interpreter) InvokeCommand(argv []string) (parserState, *TclError) {
	if len(argv) < 1 {
		i.result = "InvokeCommand called without arguments"
		return stateError, NewTclError(EILLARG, i.result)
	}
	name := argv[0]
	c, ok := i.commands[name]
	if !ok {
		i.result = fmt.Sprintf("No such command '%s'", name)
		return stateError, NewTclError(ECMDUNDEF, i.result)
	}
	s, r := c.function(i, argv, c.privdata)
	i.result = r
	return s, nil
}

// Evaluate interprets the given Tcl text.
func (i *Interpreter) Evaluate(tcl string) (parserState, *TclError) {
	// command and arguments to be invoked
	argv := make([]string, 0)
	p := NewParser(tcl)

	// TODO: with the new lexer, interpreter will need to keep track of open quotes
	//       in order to know whether to append latest token to previous or add as
	//       a new argument to the current command

	// TODO: handle escaped newline at end of string (inside both " and {, converts to space)
	for {
		prevtoken := p.token
		p.parseToken()
		if p.token == tokenEOF {
			break
		}
		t := p.GetTokenText()
		if p.token == tokenVariable {
			// Get variable value
			v, err := i.GetVariable(t)
			if err != nil {
				i.result = fmt.Sprintf("No such variable '%s'", t)
				return stateError, err
			}
			t = v

		} else if p.token == tokenCommand {
			// Evaluate command invocation
			retcode, err := i.Evaluate(t)
			if retcode != stateOK {
				return retcode, err
			}
			t = i.result

		} else if p.token == tokenEscape {
			// TODO: handle variable reference or command invocation
			// TODO: need to save parser state...
			// TODO: evaluate variable/command...
			// TODO: append result to existing string...
			// TODO: continue parsing as string

		} else if p.token == tokenSeparator {
			// Not finished parsing, continue
			continue
		}

		if p.token == tokenEOL {
			if len(argv) > 0 {
				// Parsing complete, invoke the command.
				retcode, err := i.InvokeCommand(argv)
				if retcode != stateOK {
					return retcode, err
				}
			}
			continue
		}

		// We have a new token, append to the previous or as new arg?
		if prevtoken == tokenSeparator || prevtoken == tokenEOL {
			argv = append(argv, t)
		} else {
			// interpolation
			last := len(argv) - 1
			argv[last] = argv[last] + t
		}
	}
	return stateOK, nil
}

// registerCoreCommands registers the built-in commands provided by this
// package so that they may be used by other Tcl scripts.
func (i *Interpreter) registerCoreCommands() {
	//     int j; char *name[] = {"+","-","*","/",">",">=","<","<=","==","!="};
	//     for (j = 0; j < (int)(sizeof(name)/sizeof(char*)); j++)
	//         picolRegisterCommand(i,name[j],picolCommandMath,NULL);
	i.RegisterCommand("set", commandSet, nil)
	i.RegisterCommand("puts", commandPuts, nil)
	i.RegisterCommand("if", commandIf, nil)
	// picolRegisterCommand(i,"while",picolCommandWhile,NULL);
	// picolRegisterCommand(i,"break",picolCommandRetCodes,NULL);
	// picolRegisterCommand(i,"continue",picolCommandRetCodes,NULL);
	// picolRegisterCommand(i,"proc",picolCommandProc,NULL);
	// picolRegisterCommand(i,"return",picolCommandReturn,NULL);
}

// arityError is a convenience method for commands to report an error
// with the number of arguments given to the command.
func (i *Interpreter) arityError(name string) parserState {
	i.result = fmt.Sprintf("Wrong number of arguments for %s", name)
	return stateError
}
