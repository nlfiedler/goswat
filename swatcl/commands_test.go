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
// commandBreak
//

func TestCommandBreak(t *testing.T) {
	interp := NewInterpreter()
	result := interp.Evaluate("break foo bar")
	if result.Ok() {
		t.Error("'break foo bar' is okay?")
	}
	result = interp.Evaluate("break")
	if !result.Ok() {
		t.Error("'break' not okay?")
	}
	if result.ErrorMessage() != "" {
		t.Error("'break' yielded error message")
	}
	if result.Result() != "" {
		t.Errorf("'break' yielded wrong result: %s", result.Result())
	}
	if result.ReturnCode() != returnBreak {
		t.Errorf("'break' yielded wrong code: %v", result.ReturnCode())
	}
}

//
// commandContinue
//

func TestCommandContinue(t *testing.T) {
	interp := NewInterpreter()
	result := interp.Evaluate("continue foo bar")
	if result.Ok() {
		t.Error("'continue foo bar' is okay?")
	}
	result = interp.Evaluate("continue")
	if !result.Ok() {
		t.Error("'continue' not okay?")
	}
	if result.ErrorMessage() != "" {
		t.Error("'continue' yielded error message")
	}
	if result.Result() != "" {
		t.Errorf("'continue' yielded wrong result: %s", result.Result())
	}
	if result.ReturnCode() != returnContinue {
		t.Errorf("'continue' yielded wrong code: %v", result.ReturnCode())
	}
}

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
// commandProc
//

func TestCommandProc(t *testing.T) {
	interp := NewInterpreter()
	input := `proc say_hello`
	result := interp.Evaluate(input)
	if result.Ok() {
		t.Error("proc unary should fail")
	}
	if result.ErrorCode() != EARGUMENT {
		t.Errorf("proc unary wrong error code: %d", result.ErrorCode())
	}
	if result.ErrorMessage() != "Wrong number of arguments for 'proc'" {
		t.Errorf("proc unary wrong error message: %v",
			result.ErrorMessage())
	}
	input = `proc say_hello {name} {
	return "Hello $name!"
}
say_hello "world"
`
	result = interp.Evaluate(input)
	if !result.Ok() {
		t.Errorf("proc invocation failed: %v", result.ErrorMessage())
	}
	if result.Result() != "Hello world!" {
		t.Errorf("proc invocation gave wrong result: %v", result.Result())
	}
	_, err := interp.GetVariable("name")
	if err != VariableUndefined {
		t.Error("proc invocation left variable defined")
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
// commandReturn
//

func TestCommandReturn(t *testing.T) {
	interp := NewInterpreter()
	result := interp.Evaluate("return")
	if !result.Ok() {
		t.Error("'return' not okay?")
	}
	if result.ErrorMessage() != "" {
		t.Error("'return' yielded error message")
	}
	if result.Result() != "" {
		t.Errorf("'return' yielded wrong result: %s", result.Result())
	}
	if result.ReturnCode() != returnReturn {
		t.Errorf("'return' yielded wrong code: %v", result.ReturnCode())
	}
}

func TestCommandReturnOk(t *testing.T) {
	interp := NewInterpreter()
	input := "return -code ok"
	result := interp.Evaluate(input)
	if !result.Ok() {
		t.Errorf("'%s' not okay?", input)
	}
	if result.ErrorMessage() != "" {
		t.Errorf("'%s' yielded error message: %s", input, result.ErrorMessage())
	}
	if result.Result() != "" {
		t.Errorf("'%s' yielded wrong result: %s", input, result.Result())
	}
	if result.ReturnCode() != returnOk {
		t.Errorf("'%s' yielded wrong code: %v", input, result.ReturnCode())
	}
}

func TestCommandReturnError(t *testing.T) {
	interp := NewInterpreter()
	input := "return -code fubar"
	result := interp.Evaluate(input)
	if result.Ok() {
		t.Errorf("'%s' is okay?", input)
	}
	input = "return -code error"
	result = interp.Evaluate(input)
	if result.Ok() {
		t.Errorf("'%s' is okay?", input)
	}
	if result.ErrorMessage() != "" {
		t.Errorf("'%s' yielded error message: %s", input, result.ErrorMessage())
	}
	if result.Result() != "" {
		t.Errorf("'%s' yielded wrong result: %s", input, result.Result())
	}
	if result.ReturnCode() != returnError {
		t.Errorf("'%s' yielded wrong code: %v", input, result.ReturnCode())
	}
}

func TestCommandReturnResult(t *testing.T) {
	interp := NewInterpreter()
	input := "return {a b c}"
	result := interp.Evaluate(input)
	if !result.Ok() {
		t.Errorf("'%s' not okay?", input)
	}
	if result.ErrorMessage() != "" {
		t.Errorf("'%s' yielded error message: %s", input, result.ErrorMessage())
	}
	if result.Result() != "a b c" {
		t.Errorf("'%s' yielded wrong result: %s", input, result.Result())
	}
	if result.ReturnCode() != returnReturn {
		t.Errorf("'%s' yielded wrong code: %v", input, result.ReturnCode())
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

//
// commandWhile
//

func TestCommandWhile(t *testing.T) {
	interp := NewInterpreter()
	input := "while {foo}"
	result := interp.Evaluate(input)
	if result.Ok() {
		t.Error("'while' with one argument is okay?")
	}
	if result.ErrorMessage() != "Wrong number of arguments for 'while'" {
		t.Errorf("'while' yielded wrong error message: %s", result.ErrorMessage())
	}
	input = "while {foo} {bar} {baz}"
	result = interp.Evaluate(input)
	if result.Ok() {
		t.Error("'while' with three arguments is okay?")
	}
	if result.ErrorMessage() != "Wrong number of arguments for 'while'" {
		t.Errorf("'while' yielded wrong error message: %s", result.ErrorMessage())
	}
	input = "while {[foo]} {[bar]}"
	result = interp.Evaluate(input)
	if result.Ok() {
		t.Error("'while' unknown commands is okay?")
	}
	if result.ErrorMessage() != "No such command 'foo'" {
		t.Errorf("'while' yielded error message: %s", result.ErrorMessage())
	}
	input = "while {1} {[bar]}"
	result = interp.Evaluate(input)
	if result.Ok() {
		t.Error("'while' unknown commands is okay?")
	}
	if result.ErrorMessage() != "No such command 'bar'" {
		t.Errorf("'while' yielded error message: %s", result.ErrorMessage())
	}
	input = `set x 0
while {$x < 10} {
	set x [expr $x + 1]
}
`
	result = interp.Evaluate(input)
	if !result.Ok() {
		t.Error("'while x < 10' not okay?")
	}
	if result.ErrorMessage() != "" {
		t.Errorf("'while x < 10' yielded error message: %s", result.ErrorMessage())
	}
	if result.Result() != "" {
		t.Errorf("'while x < 10' yielded wrong result: %s", result.Result())
	}
	if result.ReturnCode() != returnOk {
		t.Errorf("'while x < 10' yielded wrong code: %v", result.ReturnCode())
	}
	val, err := interp.GetVariable("x")
	if err != nil {
		t.Errorf("failed to get variable x: %s", err)
	}
	if val != "10" {
		t.Errorf("unexpected value '%s' for variable x", val)
	}
}
