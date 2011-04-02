//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"testing"
)

//
// commandSet
//

func TestCommandSet(t *testing.T) {
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

func TestCommandSet1Undef(t *testing.T) {
	interp := NewInterpreter()
	state, _ := interp.Evaluate("set foo")
	if state != stateError {
		t.Error("expected error state")
	}
	if interp.result != "Variable 'foo' undefined" {
		t.Error("expected undefined variable error")
	}
}

func TestCommandSet1(t *testing.T) {
	interp := NewInterpreter()
	err := interp.SetVariable("foo", "bar")
	state, err := interp.Evaluate("set foo")
	if err != nil {
		t.Error("failed to get variable")
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

func TestCommandSetNoFrame(t *testing.T) {
	interp := NewInterpreter()
	interp.popFrame()
	state, _ := interp.Evaluate("set foo bar")
	if state != stateError {
		t.Error("expected error state")
	}
	if interp.result != "Empty call stack, cannot set 'foo'" {
		t.Error("expected empty stack error")
	}
}

func TestCommandSet1NoFrame(t *testing.T) {
	interp := NewInterpreter()
	interp.popFrame()
	state, _ := interp.Evaluate("set foo")
	if state != stateError {
		t.Error("expected error state")
	}
	if interp.result != "Empty call stack, cannot get 'foo'" {
		t.Error("expected empty stack error")
	}
}

//
// commandIf
//

// TODO: need expr support for if and while
// func TestCommandIf(t *testing.T) {
// 	interp := NewInterpreter()
// 	state, err := interp.Evaluate("if {1} { set foo bar } else { set foo quux }")
// 	if err != nil {
// 		t.Errorf("failed to invoke command if: %s", err)
// 	}
// 	if state != stateOK {
// 		t.Errorf("if state = %d", state)
// 	}
// 	if interp.result != "bar" {
// 		t.Errorf("if result = %s", interp.result)
// 	}
// 	val, err := interp.GetVariable("foo")
// 	if err != nil {
// 		t.Errorf("failed to get variable foo: %s", err)
// 	}
// 	if val != "bar" {
// 		t.Errorf("unexpected value '%s' for variable foo", val)
// 	}
// }
