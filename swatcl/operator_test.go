//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"testing"
)

func TestOperatorUnaryPlus(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["+1"] = "1"
	values["+0"] = "0"
	values["+1.23"] = "1.23"
	values["+1234567890"] = "1234567890"
	i.SetVariable("foo", "-123")
	values["+${foo}"] = "-123"
	evaluateAndCompare(i, values, t)
}

func TestOperatorBinaryPlus(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["1 + 1"] = "2"
	values["1 + -1"] = "0"
	values["-1 + 1"] = "0"
	values["-1 + -1"] = "-2"
	values["1 + 0"] = "1"
	values["1 + 1.234"] = "2.234"
	values["1.234 + 1"] = "2.234"
	values["1.1 + 1.1"] = "2.2"
	values["1 + 1234567890"] = "1234567891"
	values["9223372036854775807 + 1"] = "-9223372036854775808"
	i.SetVariable("foo", "123")
	values["1 + ${foo}"] = "124"
	values["'foo' + 1"] = "error"
	values["\"foo\" + 1"] = "error"
	values["1 + \"foo\""] = "error"
	values["\"foo\" + \"bar\""] = "error"
	evaluateAndCompare(i, values, t)
}

func TestOperatorUnaryMinus(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["-1"] = "-1"
	values["-0"] = "0"
	values["-1.23"] = "-1.23"
	values["-1234567890"] = "-1234567890"
	i.SetVariable("foo", "123")
	values["-${foo}"] = "-123"
	evaluateAndCompare(i, values, t)
}

func TestOperatorBinaryMinus(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["1 - 1"] = "0"
	values["1 - -1"] = "2"
	values["-1 - 1"] = "-2"
	values["-1 - -1"] = "0"
	values["1.1 - 1.1"] = "0"
	values["1 - 0"] = "1"
	values["1 - 1.234"] = "-0.23399999999999999"
	values["1.234 - 1"] = "0.23399999999999999"
	values["1 - 1234567890"] = "-1234567889"
	i.SetVariable("foo", "123")
	values["1 - ${foo}"] = "-122"
	values["\"abc\" - 1"] = "error"
	values["1 - \"abc\""] = "error"
	values["\"abc\" - \"123\""] = "error"
	evaluateAndCompare(i, values, t)
}

func TestOperatorMultiply(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["1 * 1"] = "1"
	values["1 * -1"] = "-1"
	values["-1 * 1"] = "-1"
	values["-1 * -1"] = "1"
	values["101 * 100"] = "10100"
	values["10 * 4"] = "40"
	values["4 * 1.234"] = "4.936"
	values["1.234 * 4"] = "4.936"
	values["1.2 * 1.5"] = "1.7999999999999998"
	values["9223372036854775807 * 2"] = "-2"
	i.SetVariable("foo", "123")
	values["2 * ${foo}"] = "246"
	values["\"abc\" * 1"] = "error"
	values["1 * \"abc\""] = "error"
	values["\"abc\" * \"123\""] = "error"
	evaluateAndCompare(i, values, t)
}

func TestOperatorDivide(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["1 / 1"] = "1"
	values["4 / 2"] = "2"
	values["10 / 4"] = "2"
	values["-1 / 1"] = "-1"
	values["4 / -2"] = "-2"
	values["-10 / -4"] = "2"
	values["4 / 1.234"] = "3.2414910858995136"
	values["1.234 / 4"] = "0.3085"
	values["1.2 / 1.5"] = "0.7999999999999999"
	values["9223372036854775807 / 2"] = "4611686018427387903"
	i.SetVariable("foo", "123")
	values["${foo} / 10"] = "12"
	values["\"abc\" / 1"] = "error"
	values["1 / \"abc\""] = "error"
	values["\"abc\" / \"123\""] = "error"
	evaluateAndCompare(i, values, t)
}

func TestOperatorRemainder(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["1 % 1"] = "0"
	values["4 % 2"] = "0"
	values["10 % 4"] = "2"
	values["5 % 3"] = "2"
	values["5 % -3"] = "2"
	values["-5 % 3"] = "-2"
	values["-5 % -3"] = "-2"
	values["9223372036854775807 % 2"] = "1"
	i.SetVariable("foo", "123")
	values["${foo} % 5"] = "3"
	values["4 % 1.234"] = "error"
	values["1.234 % 4"] = "error"
	values["1.2 % 1.5"] = "error"
	values["\"abc\" % 1"] = "error"
	values["1 % \"abc\""] = "error"
	values["\"abc\" % \"123\""] = "error"
	evaluateAndCompare(i, values, t)
}

func TestOperatorPrecedence(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["1 + 2 * 3"] = "7"
	values["3 * 1 + 2"] = "5"
	values["(1 + 2) * 3"] = "9"
	values["3 * (1 + 2)"] = "9"
	values["((1 + 1) - 2) * 3"] = "0"
	evaluateAndCompare(i, values, t)
}

func TestOperatorStringEquality(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["{a} eq {a}"] = "1"
	values["{a} eq {b}"] = "0"
	values["{a} ne {b}"] = "1"
	values["{a} ne {a}"] = "0"
	evaluateAndCompare(i, values, t)
}

func TestOperatorLessThan(t *testing.T) {
	i := NewInterpreter()
	values := make(map[string]string)
	values["1 < 1"] = "0"
	values["4 < 2"] = "0"
	values["1 < 4"] = "1"
	values["-5 < 3"] = "1"
	values["5.0 < -3"] = "0"
	values["-5.0 < 3"] = "1"
	values["5.0 < 3.0"] = "0"
	values["3.0 < 5.0"] = "1"
	values["{abc} < {def}"] = "1"
	values["{def} < {abc}"] = "0"
	evaluateAndCompare(i, values, t)
}
