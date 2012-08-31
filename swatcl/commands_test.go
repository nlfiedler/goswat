//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"bytes"
	"testing"
)

func TestCommandExpr(t *testing.T) {
	interp := NewInterpreter()
	result, _, err := interp.Evaluate("expr 8.2 + 6")
	if err != nil {
		t.Errorf("failed to invoke command expr: %s", err)
	}
	if result != "14.2" {
		t.Errorf("expr returned wrong result: '%s'", result)
	}
}

//
// commandIf
//

func TestCommandIf(t *testing.T) {
	interp := NewInterpreter()
	// if used to set variable
	result, _, err := interp.Evaluate("if {1} { set foo bar } else { set foo quux }")
	if err != nil {
		t.Errorf("failed to invoke command if: %s", err)
	}
	if result != "bar" {
		t.Errorf("if result = %s", result)
	}
	val, err := interp.GetVariable("foo")
	if err != nil {
		t.Errorf("failed to get variable foo: %s", err)
	}
	if val != "bar" {
		t.Errorf("unexpected value '%s' for variable foo", val)
	}
	// if with complex expression
	result, _, err = interp.Evaluate("if {\"a\" eq \"a\"} { set foo baz } else { set foo qix }")
	if err != nil {
		t.Errorf("failed to invoke command if: %s", err)
	}
	if result != "baz" {
		t.Errorf("if result = %s", result)
	}
}

//
// commandPuts
//

func TestCommandPuts(t *testing.T) {
	interp := NewInterpreter()
	result, _, err := interp.Evaluate("puts")
	if err == nil {
		t.Error("expected puts with no arguments to fail")
	}
	out := new(bytes.Buffer)
	interp.SetOutput(out)
	result, _, err = interp.Evaluate("puts \"one two three\"")
	if result != "one two three" {
		t.Error("expected puts to return its input")
	}
	result = out.String()
	if result != "one two three\n" {
		t.Errorf("expected puts to print its input; got '%s'", result)
	}
	out.Reset()
	_, _, err = interp.Evaluate("puts -nonewline \"one two three\"")
	result = out.String()
	if result != "one two three" {
		t.Errorf("expected puts to print its input; got '%s'", result)
	}
}

//
// commandSet
//

func TestCommandSet(t *testing.T) {
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

func TestCommandSet1Undef(t *testing.T) {
	interp := NewInterpreter()
	_, _, err := interp.Evaluate("set foo")
	if err == nil {
		t.Error("expected error state")
	}
	if err.String() != "Variable 'foo' undefined" {
		t.Error("expected undefined variable error")
	}
}

func TestCommandSet1(t *testing.T) {
	interp := NewInterpreter()
	err := interp.SetVariable("foo", "bar")
	result, _, err := interp.Evaluate("set foo")
	if err != nil {
		t.Error("failed to get variable")
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

func TestCommandSetNoFrame(t *testing.T) {
	interp := NewInterpreter()
	interp.popFrame()
	_, _, err := interp.Evaluate("set foo bar")
	if err == nil {
		t.Error("expected error state")
	}
	if err.String() != "Empty call stack, cannot set 'foo'" {
		t.Error("expected empty stack error")
	}
}

func TestCommandSet1NoFrame(t *testing.T) {
	interp := NewInterpreter()
	interp.popFrame()
	_, _, err := interp.Evaluate("set foo")
	if err == nil {
		t.Error("expected error state")
	}
	if err.String() != "Empty call stack, cannot get 'foo'" {
		t.Error("expected empty stack error")
	}
}
