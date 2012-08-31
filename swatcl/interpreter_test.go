//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"strings"
	"testing"
)

//
// RegisterCommand
//

func TestInterpRegisterCommand(t *testing.T) {
	interp := NewInterpreter()
	err := interp.RegisterCommand("foo", nil, nil)
	if err != nil {
		t.Error("failed to register command foo")
	}
	err = interp.RegisterCommand("foo", nil, nil)
	if err == nil {
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
	if err == nil || err.Errno != ENOSTACK {
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

func testCmd(context Interpreter, argv []string, data []string) (string, returnCode, *TclError) {
	testCmdCalled = true
	testCmdArgs = strings.Join(argv[1:], ",")
	return "cmd", returnOk, nil
}

func TestInterpInvokeCommand(t *testing.T) {
	interp := NewInterpreter()
	err := interp.RegisterCommand("foo", testCmd, nil)
	if err != nil {
		t.Error("failed to register command foo")
	}
	args := make([]string, 0)
	args = append(args, "foo")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "c")
	result, _, err := interp.InvokeCommand(args)
	if err != nil {
		t.Error("failed to invoke command foo")
	}
	if !testCmdCalled {
		t.Error("InvokeCommand failed to invoke testCmd")
	}
	if testCmdArgs != "a,b,c" {
		t.Error("testCmd did not receive expected arguments")
	}
	if result != "cmd" {
		t.Error("testCmd failed to affect result of interpreter")
	}
}

//
// Evaluate
//

func TestInterpEvaluateCommand(t *testing.T) {
	interp := NewInterpreter()
	result, _, err := interp.Evaluate("set foo bar")
	if err != nil {
		t.Error("failed to invoke command set")
	}
	if result != "bar" {
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
	result, _, err := interp.Evaluate("set $foo quux")
	if err != nil {
		t.Errorf("failed to reference variable: %s", err)
	}
	if result != "quux" {
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
