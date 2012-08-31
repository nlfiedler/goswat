//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"testing"
)

func TestFunctionAbs(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["abs(1)"] = "1"
	values["abs(-1)"] = "1"
	values["abs(1.0)"] = "1"
	values["abs(-1.0)"] = "1"
	evaluateAndCompare(i, values, t)
}

func TestFunctionBool(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["bool(1)"] = "true"
	values["bool(0)"] = "false"
	values["bool(1.0)"] = "true"
	values["bool(0.0)"] = "false"
	values["bool(\"1\")"] = "true"
	values["bool(\"on\")"] = "true"
	values["bool(\"yes\")"] = "true"
	values["bool(\"true\")"] = "true"
	values["bool(\"tRUe\")"] = "true"
	values["bool(\"0\")"] = "false"
	values["bool(\"off\")"] = "false"
	values["bool(\"no\")"] = "false"
	values["bool(\"false\")"] = "false"
	values["bool(\"faLSe\")"] = "false"
	evaluateAndCompare(i, values, t)
}

func TestFunctionCeil(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["ceil(1)"] = "1"
	values["ceil(1.1)"] = "2"
	values["ceil(1.9)"] = "2"
	evaluateAndCompare(i, values, t)
}

func TestFunctionDouble(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["double(1)"] = "1"
	values["double(1.1)"] = "1.1"
	values["double(1.9)"] = "1.9"
	evaluateAndCompare(i, values, t)
}

func TestFunctionExp(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["exp(1)"] = "2.718281828459045"
	values["exp(1.1)"] = "3.0041660239464334"
	values["exp(1.9)"] = "6.6858944422792685"
	evaluateAndCompare(i, values, t)
}

func TestFunctionFloor(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["floor(1)"] = "1"
	values["floor(1.1)"] = "1"
	values["floor(1.9)"] = "1"
	values["floor(2.0)"] = "2"
	evaluateAndCompare(i, values, t)
}

func TestFunctionFmod(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["fmod(1, 0)"] = "NaN"
	values["fmod(5.0, 2)"] = "1"
	values["fmod(5, 2.0)"] = "1"
	values["fmod(5, 2)"] = "1"
	values["fmod(5.0, 2.0)"] = "1"
	evaluateAndCompare(i, values, t)
}

func TestFunctionLog(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["log(1)"] = "0"
	values["log(1.1)"] = "0.09531017980432493"
	values["log(1.9)"] = "0.6418538861723947"
	values["log(2.0)"] = "0.6931471805599453"
	evaluateAndCompare(i, values, t)
}

func TestFunctionLog10(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["log10(1)"] = "0"
	values["log10(1.1)"] = "0.04139268515822507"
	values["log10(1.9)"] = "0.2787536009528289"
	values["log10(2.0)"] = "0.3010299956639812"
	evaluateAndCompare(i, values, t)
}

func TestFunctionMax(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["max(1)"] = "1"
	values["max(1, 2, 3)"] = "3"
	values["max(100, 20, 30)"] = "100"
	values["max(1024, 512, 256, 128, 64, 32, 16, 8, 4, 2, 1)"] = "1024"
	values["max(1, 2, 3.1)"] = "3.1"
	values["max(1, 2, -3.1)"] = "2"
	evaluateAndCompare(i, values, t)
}

func TestFunctionMin(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["min(1)"] = "1"
	values["min(1, 2, 3)"] = "1"
	values["min(100, 20, 30)"] = "20"
	values["min(1024, 512, 256, 128, 64, 32, 16, 8, 4, 2, 1)"] = "1"
	values["min(1, 2, 3.1)"] = "1"
	values["min(1, 2, -3.1)"] = "-3.1"
	evaluateAndCompare(i, values, t)
}

func TestFunctionPow(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["pow(42, 0)"] = "1"
	values["pow(1, 42)"] = "1"
	values["pow(5, 2.0)"] = "25"
	values["pow(5, 2)"] = "25"
	values["pow(5, -2)"] = "0.04"
	values["pow(-5, 2)"] = "25"
	values["pow(5.0, 2.0)"] = "25"
	values["pow(-5.0, 2.1)"] = "NaN"
	evaluateAndCompare(i, values, t)
}

func TestFunctionSqrt(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["sqrt(4)"] = "2"
	values["sqrt(2.0)"] = "1.4142135623730951"
	values["sqrt(2.2)"] = "1.4832396974191326"
	values["sqrt(16.0)"] = "4"
	evaluateAndCompare(i, values, t)
}

func TestFunctionErrors(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["foobar(123)"] = "unsupported function"
	values["abs()"] = "takes exactly one argument"
	values["bool()"] = "takes exactly one argument"
	values["ceil()"] = "takes exactly one argument"
	values["double()"] = "takes exactly one argument"
	values["exp()"] = "takes exactly one argument"
	values["floor()"] = "takes exactly one argument"
	values["log()"] = "takes exactly one argument"
	values["log10()"] = "takes exactly one argument"
	values["max()"] = "takes at least one argument"
	values["min()"] = "takes at least one argument"
	values["rand(1)"] = "takes no arguments"
	values["sqrt()"] = "takes exactly one argument"
	values["srand()"] = "takes exactly one argument"
	values["abs(1, 2)"] = "takes exactly one argument"
	values["bool(1, 2)"] = "takes exactly one argument"
	values["ceil(1, 2)"] = "takes exactly one argument"
	values["double(1, 2)"] = "takes exactly one argument"
	values["exp(1, 2)"] = "takes exactly one argument"
	values["floor(1, 2)"] = "takes exactly one argument"
	values["log(1, 2)"] = "takes exactly one argument"
	values["log10(1, 2)"] = "takes exactly one argument"
	values["sqrt(1, 2)"] = "takes exactly one argument"
	values["srand(1, 2)"] = "takes exactly one argument"
	values["abs({a})"] = "takes only ints and floats"
	values["bool({a})"] = "expected 'string is boolean' value"
	values["ceil({a})"] = "takes only ints and floats"
	values["double({a})"] = "takes only ints and floats"
	values["exp({a})"] = "takes only ints and floats"
	values["floor({a})"] = "takes only ints and floats"
	values["log({a})"] = "takes only ints and floats"
	values["log10({a})"] = "takes only ints and floats"
	values["max({a}, {b})"] = "takes only ints and floats"
	values["min({a}, {b})"] = "takes only ints and floats"
	values["sqrt({a})"] = "takes only ints and floats"
	values["srand({a})"] = "takes only integers"
	evaluateForError(i, values, t)
}
