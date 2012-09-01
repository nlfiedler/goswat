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

// evaluateAndCompare invokes EvaluateExpession on each of the map keys and
// compares the result to the corresponding map value.
func evaluateAndCompare(interp Interpreter, values map[string]string, t *testing.T) {
	eval := newEvaluator(interp)
	for k, v := range values {
		r := eval.Evaluate(k)
		if !r.Ok() {
			if v != "error" {
				t.Errorf("evaluation of '%s' failed: %s", k, r.Error())
			}
		} else if r.Result() != v {
			t.Errorf("evaluation of '%s' resulted in '%s'", k, r)
		}
	}
}

// evaluateForError invokes EvaluateExpession on each of the map keys and
// compares the (expected) error result to the corresponding map value.
func evaluateForError(interp Interpreter, values map[string]string, t *testing.T) {
	eval := newEvaluator(interp)
	for k, v := range values {
		r := eval.Evaluate(k)
		if r.Ok() {
			t.Errorf("evaluation of '%s' should have faild with '%s'", k, v)
		} else if !strings.Contains(r.Error(), v) {
			t.Errorf("evaluation of '%s' yielded wrong error: '%s'", k, r)
		}
	}
}

func TestExprInteger(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	// force coercion of numbers by using operators
	values["0 + 123"] = "123"
	values["0 + 0xcafebabe"] = "3405691582"
	values["0 + 0126547"] = "44391"
	evaluateAndCompare(i, values, t)
}

func TestExprFloat(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	// force coercion of numbers by using operators
	values["0 + 1.23"] = "1.23"
	values["0 + 3."] = "3"
	values["0 + 0.0001"] = "0.0001"
	values["0 + 6E4"] = "60000"
	values["0 + 7.91e+16"] = "7.91e+16"
	values["0 + 1e+012"] = "1e+12"
	values["0 + 4.4408920985006262e-016"] = "4.440892098500626e-16"
	evaluateAndCompare(i, values, t)
}

func TestExprString(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["\"123\""] = "123"
	values["\"0xcafebabe\""] = "0xcafebabe"
	values["\"foobar\""] = "foobar"
	values["\"foo\\nbar\""] = "foo\nbar"
	values["{foobar}"] = "foobar"
	values["{foo\\nbar}"] = "foo\\nbar"
	evaluateAndCompare(i, values, t)
}

func TestExprMissingParen(t *testing.T) {
	i := NewInterpreter()
	eval := newEvaluator(i)
	r := eval.Evaluate("(1 + 2")
	if r.Ok() {
		t.Error("expected missing close paren to fail")
	}
	r = eval.Evaluate("1 + 2)")
	if r.Ok() {
		t.Error("expected missing open paren to fail")
	}
}

func TestExprNestedCmd(t *testing.T) {
	i := NewInterpreter()
	eval := newEvaluator(i)
	r := eval.Evaluate("[expr 8.2 + 6]")
	if !r.Ok() {
		t.Errorf("evaluating nested [expr] failed: %s", r.Error())
	}
	if r.Result() != "14.2" {
		t.Errorf("nested [expr] returned wrong result: '%s'", r)
	}
}

func TestExprVariable(t *testing.T) {
	i := NewInterpreter()
	i.SetVariable("foo", "8.2")
	eval := newEvaluator(i)
	r := eval.Evaluate("[expr $foo + 6]")
	if !r.Ok() {
		t.Errorf("evaluating nested [expr] failed: %s", r.Error())
	}
	if r.Result() != "14.2" {
		t.Errorf("nested [expr] returned wrong result: '%s'", r)
	}
}
