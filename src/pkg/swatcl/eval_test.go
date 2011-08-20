//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"testing"
)

func assertNoError(t *testing.T, err *TclError) {
	if err != nil {
		t.Errorf("error in addition: %s", err)
	}
}

func assertError(t *testing.T, err *TclError) {
	if err == nil {
		t.Error("expected an error")
	}
}

// evalStrAndCompare invokes evalString on each of the map keys and
// compares the result to the corresponding map value.
func evalStrAndCompare(values map[string]string, t *testing.T) {
	for k, v := range values {
		r, e := evalString(k)
		if e != nil {
			t.Errorf("evaluation of '%s' failed: %s", k, e)
		}
		if r != v {
			t.Errorf("evaluation of '%s' resulted in '%s'", k, r)
		}
	}
}

func TestEvalBoolean(t *testing.T) {
	badbools := [...]string{"foo", "1.0", "sure", "yesarooney", "no, sir"}
	for bad := range badbools {
		_, err := evalBoolean(badbools[bad])
		if err == nil || err.Errno != EBADBOOL {
			t.Error("expected bad boolean error")
		}
	}
	tests := make(map[string]bool)
	tests["FaLse"] = false
	tests["tRUE"] = true
	tests["No"] = false
	tests["yeS"] = true
	tests["0"] = false
	tests["1"] = true
	tests["10"] = true
	for k, v := range tests {
		b, err := evalBoolean(k)
		if err != nil {
			t.Errorf("unexpected error in evalBoolean: %s", err)
		}
		if b != v {
			t.Errorf("expected %t for %s", v, k)
		}
	}
}

func TestEvalString(t *testing.T) {
	values := make(map[string]string)
	values["abc"] = "abc"
	values["abc\\adef"] = "abc\adef"
	values["abc\\bdef"] = "abc\bdef"
	values["abc\\fdef"] = "abc\fdef"
	values["abc\\ndef"] = "abc\ndef"
	values["abc\\rdef"] = "abc\rdef"
	values["abc\\tdef"] = "abc\tdef"
	values["abc\\vdef"] = "abc\vdef"
	values["foo\\u005cbar"] = "foo\\bar"
	values["foo\\x5cbar"] = "foo\\bar"
	values["foo\\043bar"] = "foo#bar"
	evalStrAndCompare(values, t)
}

func TestCoerceNumber(t *testing.T) {
	ints := make(map[string]int64)
	ints["1"] = 1
	ints["123"] = 123
	ints["+123"] = 123
	ints["-123"] = -123
	for k, v := range ints {
		n, err := coerceNumber(k)
		if err != nil {
			t.Errorf("unexpected error in coerceNumber: %s", err)
		}
		if n != v {
			t.Errorf("expected %v (%T) for %s, but got %v (%T)", v, v, k, n, n)
		}
	}

	floats := make(map[string]float64)
	floats[".01"] = 0.01
	floats["1.0"] = 1.0
	floats["123.1"] = 123.1
	floats["+123.1"] = 123.1
	floats["-123.1"] = -123.1
	for k, v := range floats {
		n, err := coerceNumber(k)
		if err != nil {
			t.Errorf("unexpected error in coerceNumber: %s", err)
		}
		if n != v {
			t.Errorf("expected %v (%T) for %s, but got %v (%T)", v, v, k, n, n)
		}
	}

	strings := [...]string{"a.10", "!@#$", "foo"}
	for i := range strings {
		s := strings[i]
		n, err := coerceNumber(s)
		if err != nil {
			t.Errorf("unexpected error in coerceNumber: %s", err)
		}
		if n != s {
			t.Errorf("expected %v (%T) for %s, but got %v (%T)", s, s, s, n, n)
		}
	}
}

func TestPerformAddition(t *testing.T) {
	v, err := performAddition(int64(1), int64(1))
	assertNoError(t, err)
	if v != int64(2) {
		t.Errorf("1 + 1 = %v ?!?", v)
	}
	v, err = performAddition(float64(1.234), int64(1))
	assertNoError(t, err)
	if v != float64(2.234) {
		t.Errorf("1.234 + 1 = %v ?!?", v)
	}
	v, err = performAddition(int64(1), float64(1.234))
	assertNoError(t, err)
	if v != float64(2.234) {
		t.Errorf("1 + 1.234 = %v ?!?", v)
	}
	v, err = performAddition(float64(1.1), float64(1.1))
	assertNoError(t, err)
	if v != float64(2.2) {
		t.Errorf("1.1 + 1.1 = %v ?!?", v)
	}
	v, err = performAddition(int64(9223372036854775807), int64(1))
	assertNoError(t, err)
	if v != int64(-9223372036854775808) {
		t.Errorf("9223372036854775807 + 1 = %v ?!?", v)
	}
	_, err = performAddition("abc", int64(1))
	assertError(t, err)
	_, err = performAddition(int64(1), "abc")
	assertError(t, err)
	_, err = performAddition("abc", "123")
	assertError(t, err)
}

func TestPerformSubtraction(t *testing.T) {
	v, err := performSubtraction(int64(1), int64(1))
	assertNoError(t, err)
	if v != int64(0) {
		t.Errorf("1 - 1 = %v ?!?", v)
	}
	v, err = performSubtraction(float64(1.234), int64(1))
	assertNoError(t, err)
	if v != float64(0.23399999999999999) {
		t.Errorf("1.234 - 1 = %v ?!?", v)
	}
	v, err = performSubtraction(int64(1), float64(1.234))
	assertNoError(t, err)
	if v != float64(-0.23399999999999999) {
		t.Errorf("1 - 1.234 = %v ?!?", v)
	}
	v, err = performSubtraction(float64(1.1), float64(1.1))
	assertNoError(t, err)
	if v != float64(0) {
		t.Errorf("1.1 - 1.1 = %v ?!?", v)
	}
	_, err = performSubtraction("abc", int64(1))
	assertError(t, err)
	_, err = performSubtraction(int64(1), "abc")
	assertError(t, err)
	_, err = performSubtraction("abc", "123")
	assertError(t, err)
}

func TestPerformMultiplication(t *testing.T) {
	v, err := performMultiplication(int64(1), int64(1))
	assertNoError(t, err)
	if v != int64(1) {
		t.Errorf("1 * 1 = %v ?!?", v)
	}
	v, err = performMultiplication(int64(101), int64(100))
	assertNoError(t, err)
	if v != int64(10100) {
		t.Errorf("101 * 100 = %v ?!?", v)
	}
	v, err = performMultiplication(float64(1.234), int64(4))
	assertNoError(t, err)
	if v != float64(4.936) {
		t.Errorf("1.234 * 4 = %v ?!?", v)
	}
	v, err = performMultiplication(int64(4), float64(1.234))
	assertNoError(t, err)
	if v != float64(4.936) {
		t.Errorf("4 * 1.234 = %v ?!?", v)
	}
	v, err = performMultiplication(float64(1.2), float64(1.5))
	assertNoError(t, err)
	if v != float64(1.7999999999999998) {
		t.Errorf("1.2 * 1.5 = %v ?!?", v)
	}
	v, err = performMultiplication(int64(9223372036854775807), int64(2))
	assertNoError(t, err)
	if v != int64(-2) {
		t.Errorf("9223372036854775807 * 2 = %v ?!?", v)
	}
	_, err = performMultiplication("abc", int64(1))
	assertError(t, err)
	_, err = performMultiplication(int64(1), "abc")
	assertError(t, err)
	_, err = performMultiplication("abc", "123")
	assertError(t, err)
}

func TestPerformDivision(t *testing.T) {
	v, err := performDivision(int64(4), int64(2))
	assertNoError(t, err)
	if v != int64(2) {
		t.Errorf("4 / 2 = %v ?!?", v)
	}
	v, err = performDivision(int64(100), int64(100))
	assertNoError(t, err)
	if v != int64(1) {
		t.Errorf("100 / 100 = %v ?!?", v)
	}
	v, err = performDivision(float64(1.234), int64(4))
	assertNoError(t, err)
	if v != float64(0.3085) {
		t.Errorf("1.234 / 4 = %v ?!?", v)
	}
	v, err = performDivision(int64(4), float64(1.234))
	assertNoError(t, err)
	if v != float64(3.2414910858995136) {
		t.Errorf("4 / 1.234 = %v ?!?", v)
	}
	v, err = performDivision(float64(1.2), float64(1.5))
	assertNoError(t, err)
	if v != float64(0.7999999999999999) {
		t.Errorf("1.2 / 1.5 = %v ?!?", v)
	}
	v, err = performDivision(int64(9223372036854775807), int64(2))
	assertNoError(t, err)
	if v != int64(4611686018427387903) {
		t.Errorf("9223372036854775807 / 2 = %v ?!?", v)
	}
	_, err = performDivision("abc", int64(1))
	assertError(t, err)
	_, err = performDivision(int64(1), "abc")
	assertError(t, err)
	_, err = performDivision("abc", "123")
	assertError(t, err)
}

func TestPerformRemainder(t *testing.T) {
	v, err := performRemainder(int64(4), int64(2))
	assertNoError(t, err)
	if v != int64(0) {
		t.Errorf("4 %% 2 = %v ?!?", v)
	}
	v, err = performRemainder(int64(5), int64(3))
	assertNoError(t, err)
	if v != int64(2) {
		t.Errorf("5 %% 3 = %v ?!?", v)
	}
	v, err = performRemainder(int64(-5), int64(3))
	assertNoError(t, err)
	if v != int64(-2) {
		t.Errorf("-5 %% 3 = %v ?!?", v)
	}
	v, err = performRemainder(int64(-5), int64(-3))
	assertNoError(t, err)
	if v != int64(-2) {
		t.Errorf("-5 %% -3 = %v ?!?", v)
	}
	v, err = performRemainder(int64(9223372036854775807), int64(2))
	assertNoError(t, err)
	if v != int64(1) {
		t.Errorf("9223372036854775807 %% 2 = %v ?!?", v)
	}
	_, err = performRemainder(float64(1.234), int64(4))
	assertError(t, err)
	_, err = performRemainder(int64(4), float64(1.234))
	assertError(t, err)
	_, err = performRemainder("abc", int64(1))
	assertError(t, err)
	_, err = performRemainder(int64(1), "abc")
	assertError(t, err)
	_, err = performRemainder("abc", "123")
	assertError(t, err)
}
