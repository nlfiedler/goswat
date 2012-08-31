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
		r, _, e := eval.Evaluate(k)
		if e != nil {
			if v != "error" {
				t.Errorf("evaluation of '%s' failed: %s", k, e)
			}
		} else if r != v {
			t.Errorf("evaluation of '%s' resulted in '%s'", k, r)
		}
	}
}

// evaluateForError invokes EvaluateExpession on each of the map keys and
// compares the (expected) error result to the corresponding map value.
func evaluateForError(interp Interpreter, values map[string]string, t *testing.T) {
	eval := newEvaluator(interp)
	for k, v := range values {
		_, _, e := eval.Evaluate(k)
		if e == nil {
			t.Errorf("evaluation of '%s' should have faild with '%s'", k, v)
		} else if !strings.Contains(e.String(), v) {
			t.Errorf("evaluation of '%s' yielded wrong error: '%s'", k, e)
		}
	}
}

func TestExprInteger(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["123"] = "123"
	values["0xcafebabe"] = "3405691582"
	values["0126547"] = "44391"
	evaluateAndCompare(i, values, t)
}

func TestExprFloat(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["1.23"] = "1.23"
	values["3."] = "3"
	values["0.0001"] = "0.0001"
	values["6E4"] = "60000"
	values["7.91e+16"] = "7.91e+16"
	values["1e+012"] = "1e+12"
	values["4.4408920985006262e-016"] = "4.440892098500626e-16"
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
	_, _, err := eval.Evaluate("(1 + 2")
	if err == nil {
		t.Error("expected missing close paren to fail")
	}
	_, _, err = eval.Evaluate("1 + 2)")
	if err == nil {
		t.Error("expected missing open paren to fail")
	}
}

func TestExprNestedCmd(t *testing.T) {
	i := NewInterpreter()
	eval := newEvaluator(i)
	result, _, err := eval.Evaluate("[expr 8.2 + 6]")
	if err != nil {
		t.Errorf("evaluating nested [expr] failed: %s", err)
	}
	if result != "14.2" {
		t.Errorf("nested [expr] returned wrong result: '%s'", result)
	}
}

func TestExprVariable(t *testing.T) {
	i := NewInterpreter()
	i.SetVariable("foo", "8.2")
	eval := newEvaluator(i)
	result, _, err := eval.Evaluate("[expr $foo + 6]")
	if err != nil {
		t.Errorf("evaluating nested [expr] failed: %s", err)
	}
	if result != "14.2" {
		t.Errorf("nested [expr] returned wrong result: '%s'", result)
	}
}
