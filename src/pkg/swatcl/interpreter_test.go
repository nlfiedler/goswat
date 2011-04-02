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
	level    int
	frames   []callFrame
	commands map[string]swatclCmd
	result   string
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
	if result.frames != nil {
		if len(interp.frames) != len(result.frames) {
			t.Errorf("frame count did not match (expected %d, actual %d)",
				len(result.frames), len(interp.frames))
		}
		// TODO: add more frame checks
	}
	if result.commands != nil {
		if len(interp.commands) != len(result.commands) {
			t.Errorf("command count did not match (expected %d, actual %d)",
				len(result.commands), len(interp.commands))
		}
		// TODO: add more command checks
	}
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
	result := interpResult{0, make([]callFrame, 1), nil, ""}
	validateInterpreter(interp, result, t)
}

//
// popFrame
//

func TestInterpPopFrame(t *testing.T) {
	interp := NewInterpreter()
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
	state, err := interp.RegisterCommand("foo", nil, nil)
	if state != stateError {
		t.Error("should have failed to register command foo second time")
	}
	if err.Errno != ECMDDEF {
		t.Error("expected command already defined error")
	}
}

//
// GetVariable/SetVariable
//

func TestInterpGetVariableNoStack(t *testing.T) {
	interp := NewInterpreter()
	interp.popFrame()
	_, err := interp.GetVariable("foo")
	if err.Errno != ENOSTACK {
		t.Error("expected no stack error")
	}
}

func TestInterpSetVariableNoStack(t *testing.T) {
	interp := NewInterpreter()
	interp.popFrame()
	err := interp.SetVariable("foo", "bar")
	if err == nil ||  err.Errno != ENOSTACK {
		t.Error("expected no stack error in SetVariable")
	}
}

func TestInterpUndefVariable(t *testing.T) {
	interp := NewInterpreter()
	_, err := interp.GetVariable("foo")
	if err == nil || err.Errno != EVARUNDEF {
		t.Error("should have failed to get undefined variable")
	}
}

func TestInterpSetGetVariables(t *testing.T) {
	interp := NewInterpreter()
	err := interp.SetVariable("foo", "bar")
	if err != nil {
		t.Error("failed to set variable foo")
	}
	val, err := interp.GetVariable("foo")
	if err != nil {
		t.Error("failed to get variable foo")
	}
	if val != "bar" {
		t.Errorf("unexpected value '%s' for variable foo", val)
	}
}

func TestInterpSetGetFrames(t *testing.T) {
	interp := NewInterpreter()
	err := interp.SetVariable("foo", "bar")
	if err != nil {
		t.Error("failed to set variable foo")
	}
	val, err := interp.GetVariable("foo")
	if err != nil {
		t.Error("failed to get variable foo")
	}
	if val != "bar" {
		t.Errorf("unexpected value '%s' for variable foo", val)
	}
	// add a frame, set the same variable
	interp.addFrame()
	err = interp.SetVariable("foo", "quux")
	if err != nil {
		t.Error("failed to set variable foo")
	}
	val, err = interp.GetVariable("foo")
	if err != nil {
		t.Error("failed to get variable foo")
	}
	if val != "quux" {
		t.Errorf("unexpected value '%s' for variable foo", val)
	}
	// remove the frame, check original variable
	interp.popFrame()
	err = interp.SetVariable("foo", "bar")
	if err != nil {
		t.Error("failed to set variable foo")
	}
	val, err = interp.GetVariable("foo")
	if err != nil {
		t.Error("failed to get variable foo")
	}
	if val != "bar" {
		t.Errorf("unexpected value '%s' for variable foo", val)
	}
}

//
// InvokeCommand
//

var testCmdCalled bool
var testCmdArgs string

func testCmd(context *Interpreter, argv []string, data []string) (parserState, string) {
	testCmdCalled = true
	testCmdArgs = strings.Join(argv[1:], ",")
	return stateOK, "cmd"
}

func TestInterpInvokeCommand(t *testing.T) {
	interp := NewInterpreter()
	state, _ := interp.RegisterCommand("foo", testCmd, nil)
	if state != stateOK {
		t.Error("failed to register command foo")
	}
	args := make([]string, 0)
	args = append(args, "foo")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "c")
	state, err := interp.InvokeCommand(args)
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
	if interp.result != "cmd" {
		t.Error("testCmd failed to affect result of interpreter")
	}
}

//
// Evaluate
//

func TestInterpEvaluateCommand(t *testing.T) {
	interp := NewInterpreter()
	state, err := interp.Evaluate("set foo bar")
	if err != nil {
		t.Error("failed to invoke command set")
	}
	if state != stateOK {
		t.Error("command set failed to return stateOK")
	}
	if interp.result != "bar" {
		t.Error("set failed to affect result of interpreter")
	}
	val, err := interp.GetVariable("foo")
	if err != nil {
		t.Error("failed to get variable foo")
	}
	if val != "bar" {
		t.Errorf("unexpected value '%s' for variable foo", val)
	}
}

func TestInterpEvaluateVariable(t *testing.T) {
	interp := NewInterpreter()
	err := interp.SetVariable("foo", "bar")
	if err != nil {
		t.Error("failed to set variable")
	}
	state, err := interp.Evaluate("set $foo quux")
	if err != nil {
		t.Errorf("failed to reference variable: %s", err)
	}
	if state != stateOK {
		t.Error("variable reference failed to return stateOK")
	}
	if interp.result != "quux" {
		t.Error("var ref failed to affect result of interpreter")
	}
	val, err := interp.GetVariable("bar")
	if err != nil {
		t.Error("failed to get variable bar")
	}
	if val != "quux" {
		t.Errorf("unexpected value '%s' for variable bar", val)
	}
}

// TODO: test evaluate nested command
// TODO: test evaluate braced string
// TODO: test evaluate quoted string

func TestInterpParseBoolean(t *testing.T) {
	interp := NewInterpreter()
	badbools := [...]string{"foo", "1.0", "sure"}
	for bad := range badbools {
		_, err := interp.parseBoolean(badbools[bad])
		if err == nil || err.Errno != EBADBOOL {
			t.Error("expected bad boolean error")
		}
	}
	tests := make(map[string]bool)
	tests["FaLse"] = false
	tests["tRUE"] = true
	tests["No"] = false
	tests["yeS"] = true
	tests["0"] = false
	tests["1"] = true
	tests["10"] = true
	for k, v := range tests {
		b, err := interp.parseBoolean(k)
		if err != nil {
			t.Errorf("unexpected error in parseBoolean: %s", err)
		}
		if b != v {
			t.Errorf("expected %t for %s", v, k)
		}
	}
}
