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
			t.Errorf("evaluation of '%s' failed: %s", k, e)
		}
		if r != v {
			t.Errorf("evaluation of '%s' resulted in '%s'", k, r)
		}
	}
}

func TestExprInteger(t *testing.T) {
	values := make(map[string]string)
	values["123"] = "123"
	values["0xcafebabe"] = "3405691582"
	values["0126547"] = "44391"
	i := NewInterpreter()
	evaluateAndCompare(i, values, t)
}

func TestExprFloat(t *testing.T) {
	values := make(map[string]string)
	values["1.23"] = "1.23"
	values["3."] = "3"
	values["0.0001"] = "0.0001"
	values["6E4"] = "60000"
	values["7.91e+16"] = "7.91e+16"
	values["1e+012"] = "1e+12"
	values["4.4408920985006262e-016"] = "4.440892098500626e-16"
	i := NewInterpreter()
	evaluateAndCompare(i, values, t)
}

func TestExprString(t *testing.T) {
	values := make(map[string]string)
	values["\"123\""] = "123"
	values["\"0xcafebabe\""] = "0xcafebabe"
	values["\"foobar\""] = "foobar"
	values["\"foo\\nbar\""] = "foo\nbar"
	values["{foobar}"] = "foobar"
	values["{foo\\nbar}"] = "foo\\nbar"
	i := NewInterpreter()
	evaluateAndCompare(i, values, t)
}

func TestExprUnaryPlus(t *testing.T) {
	values := make(map[string]string)
	values["+1"] = "1"
	values["+0"] = "0"
	values["+1.23"] = "1.23"
	values["+1234567890"] = "1234567890"
	i := NewInterpreter()
	i.SetVariable("foo", "-123")
	values["+${foo}"] = "-123"
	evaluateAndCompare(i, values, t)
}

func TestExprBinaryPlus(t *testing.T) {
	values := make(map[string]string)
	values["1+1"] = "2"
	values["1+0"] = "1"
	values["1+1.23"] = "2.23"
	values["1+1234567890"] = "1234567891"
	i := NewInterpreter()
	i.SetVariable("foo", "123")
	values["1+${foo}"] = "124"
	evaluateAndCompare(i, values, t)
}

func TestExprUnaryMinus(t *testing.T) {
	values := make(map[string]string)
	values["-1"] = "-1"
	values["-0"] = "0"
	values["-1.23"] = "-1.23"
	values["-1234567890"] = "-1234567890"
	i := NewInterpreter()
	i.SetVariable("foo", "123")
	values["-${foo}"] = "-123"
	evaluateAndCompare(i, values, t)
}

func TestExprBinaryMinus(t *testing.T) {
	values := make(map[string]string)
	values["1-1"] = "0"
	values["1-0"] = "1"
	values["1-1.23"] = "-0.22999999999999998"
	values["1-1234567890"] = "-1234567889"
	i := NewInterpreter()
	i.SetVariable("foo", "123")
	values["1-${foo}"] = "-122"
	evaluateAndCompare(i, values, t)
}
