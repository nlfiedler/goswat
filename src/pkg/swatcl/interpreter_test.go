//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"strings"
	"testing"
)

// interpResult is the expected state of the interpreter.
type interpResult struct {
	level int
	frames []callFrame
	commands map[string]swatclCmd
	result string
}

// validateInterpreter checks that the interpreter is in exactly the
// state specified by the result, failing if that is not the case.
func validateInterpreter(interp *Interpreter, result interpResult, t *testing.T) {
	if interp.level != result.level {
		t.Errorf("level did not match (expected %d, actual %d)", result.level, interp.level)
	}
	if interp.result != result.result {
		t.Errorf("result did not match (expected %s, actual %s)", result.result, interp.result)
	}
	if len(interp.frames) != len(result.frames) {
		t.Errorf("frame count did not match (expected %d, actual %d)",
			len(result.frames), len(interp.frames))
	}
	if len(interp.commands) != len(result.commands) {
		t.Errorf("command count did not match (expected %d, actual %d)",
			len(result.commands), len(interp.commands))
	}
	// TODO: compare the commands
	// TODO: compare the frames and their contents
}

//
// NewInterpreter
//

func TestNewInterpreter(t *testing.T) {
	interp := NewInterpreter()
	result := interpResult{0, nil, nil, ""}
	validateInterpreter(interp, result, t)
}

//
// addFrame
//

func TestInterpAddFrame(t *testing.T) {
	interp := NewInterpreter()
	interp.addFrame()
	result := interpResult{0, make([]callFrame, 1), nil, ""}
	validateInterpreter(interp, result, t)
}

//
// popFrame
//

func TestInterpPopFrame(t *testing.T) {
	interp := NewInterpreter()
	interp.addFrame()
	result := interpResult{0, make([]callFrame, 1), nil, ""}
	validateInterpreter(interp, result, t)
	interp.popFrame()
	result = interpResult{0, make([]callFrame, 0), nil, ""}
	validateInterpreter(interp, result, t)
}

//
// RegisterCommand
//

func TestInterpRegisterCommand(t *testing.T) {
	interp := NewInterpreter()
	state, _ := interp.RegisterCommand("foo", nil, nil)
	if state != stateOK {
		t.Error("failed to register command foo")
	}
	cmds := make(map[string]swatclCmd, 0)
	cmds["foo"] = swatclCmd{nil, nil}
	result := interpResult{0, nil, cmds, ""}
	validateInterpreter(interp, result, t)
	state, err := interp.RegisterCommand("foo", nil, nil)
	if state != stateError {
		t.Error("should have failed to register command foo second time")
	}
	if err.Errno != ECMDDEF {
		t.Error("expected command already defined error")
	}
	result = interpResult{0, nil, cmds, "Command 'foo' already defined"}
	validateInterpreter(interp, result, t)
}

//
// GetVariable/SetVariable
//

func TestInterpGetVariableNoStack(t *testing.T) {
	interp := NewInterpreter()
	_, err := interp.GetVariable("foo")
	if err.Errno != ENOSTACK {
		t.Error("expected no stack error")
	}
}

func TestInterpSetVariableNoStack(t *testing.T) {
	interp := NewInterpreter()
	err := interp.SetVariable("foo", "bar")
	if err.Errno != ENOSTACK {
		t.Error("expected no stack error")
	}
}

func TestInterpUndefVariable(t *testing.T) {
	interp := NewInterpreter()
	interp.addFrame()
	_, err := interp.GetVariable("foo")
	if err.Errno != EVARUNDEF {
		t.Error("should have failed to get undefined variable")
	}
}

func TestInterpSetGetVariables(t *testing.T) {
	interp := NewInterpreter()
	interp.addFrame()
	err := interp.SetVariable("foo", "bar")
	if err != nil {
		t.Error("failed to set variable foo")
	}
	s, err := interp.GetVariable("foo")
	if err != nil {
		t.Error("failed to get variable foo")
	}
	if s != "bar" {
		t.Errorf("unexpected value '%s' for variable foo", s)
	}
}

//
// InvokeCommand
//

var testCmdCalled bool
var testCmdArgs string

func testCmd(context *Interpreter, argv []string, data []byte) (parserState) {
	testCmdCalled = true
	testCmdArgs = strings.Join(argv, ",")
	return stateOK
}

func TestInterpInvokeCommand(t *testing.T) {
	interp := NewInterpreter()
	state, _ := interp.RegisterCommand("foo", testCmd, nil)
	if state != stateOK {
		t.Error("failed to register command foo")
	}
	args := make([]string, 0)
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "c")
	state, err := interp.InvokeCommand("foo", args)
	if err != nil {
		t.Error("failed to invoke command foo")
	}
	if state != stateOK {
		t.Error("command foo failed to return stateOK")
	}
	if !testCmdCalled {
		t.Error("InvokeCommand failed to invoke testCmd")
	}
	if testCmdArgs != "a,b,c" {
		t.Error("testCmd did not receive expected arguments")
	}
}
