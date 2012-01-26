//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"testing"
)

func TestExprNewEvaluator(t *testing.T) {
	i := NewInterpreter()
	e := newEvaluator(i)
	if e.state != searchArgument {
		t.Error("new evaluator in incorrect state")
	}
}

// evaluateAndCompare invokes EvaluateExpession on each of the map
// keys and compares the result to the corresponding map value.
func evaluateAndCompare(interp *Interpreter, values map[string]string, t *testing.T) {
	for k, v := range values {
		r, e := EvaluateExpression(interp, k)
		if e != nil {
			if v != "error" {
				t.Errorf("evaluation of '%s' failed: %s", k, e)
			}
		} else if r != v {
			t.Errorf("evaluation of '%s' resulted in '%s'", k, r)
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
