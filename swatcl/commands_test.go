//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"bytes"
	"strings"
	"testing"
)

//
// commandExpr
//

func TestCommandExpr(t *testing.T) {
	interp := NewInterpreter()
	result := interp.Evaluate("expr 8.2 + 6")
	if !result.Ok() {
		t.Errorf("failed to invoke command expr: %s", result.Error())
	}
	if result.Result() != "14.2" {
		t.Errorf("expr returned wrong result: '%v'", result)
	}
}

//
// commandIf
//

func TestCommandIf(t *testing.T) {
	interp := NewInterpreter()
	// if used to set variable
	result := interp.Evaluate("if {1} { set foo bar } else { set foo quux }")
	if !result.Ok() {
		t.Errorf("failed to invoke command if: %s", result)
	}
	if result.Result() != "bar" {
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
	result = interp.Evaluate("if {\"a\" eq \"a\"} { set foo baz } else { set foo qix }")
	if !result.Ok() {
		t.Errorf("failed to invoke command if: %s", err)
	}
	if result.Result() != "baz" {
		t.Errorf("if result = %s", result)
	}
}

//
// commandPuts
//

func TestCommandPuts(t *testing.T) {
	interp := NewInterpreter()
	result := interp.Evaluate("puts")
	if !result.Ok() {
		t.Error("puts with no arguments should silently return")
	}
	out := new(bytes.Buffer)
	interp.SetOutput(out)
	result = interp.Evaluate("puts \"one two three\"")
	if result.Result() != "one two three" {
		t.Error("expected puts to return its input")
	}
	if out.String() != "one two three\n" {
		t.Errorf("expected puts to print its input; got '%s'", result)
	}
	out.Reset()
	result = interp.Evaluate("puts -nonewline \"one two three\"")
	if !result.Ok() {
		t.Error("unexpected error with puts -nonewline")
	}
	if out.String() != "one two three" {
		t.Errorf("expected puts to print its input; got '%s'", result)
	}
}

//
// commandSet
//

func TestCommandSet(t *testing.T) {
	interp := NewInterpreter()
	result := interp.Evaluate("set foo bar")
	if !result.Ok() {
		t.Error("failed to invoke command set")
	}
	if result.Result() != "bar" {
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
	result := interp.Evaluate("set foo")
	if result.Ok() {
		t.Error("expected error state")
	}
	if !strings.Contains(result.Error(), "variable undefined") {
		t.Error("expected undefined variable error")
	}
}

func TestCommandSet1(t *testing.T) {
	interp := NewInterpreter()
	err := interp.SetVariable("foo", "bar")
	result := interp.Evaluate("set foo")
	if !result.Ok() {
		t.Error("failed to get variable")
	}
	if result.Result() != "bar" {
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
	result := interp.Evaluate("set foo bar")
	if result.Ok() {
		t.Error("expected error state")
	}
	if !strings.Contains(result.Error(), "call stack is empty") {
		t.Error("expected empty stack error")
	}
}

func TestCommandSet1NoFrame(t *testing.T) {
	interp := NewInterpreter()
	interp.popFrame()
	result := interp.Evaluate("set foo")
	if result.Ok() {
		t.Error("expected error state")
	}
	if !strings.Contains(result.Error(), "call stack is empty") {
		t.Error("expected empty stack error")
	}
}
