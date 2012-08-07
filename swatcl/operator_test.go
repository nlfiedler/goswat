//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"testing"
)

func TestExprUnaryPlus(t *testing.T) {
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

func TestExprBinaryPlus(t *testing.T) {
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

func TestExprUnaryMinus(t *testing.T) {
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

func TestExprBinaryMinus(t *testing.T) {
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

func TestExprBinaryMultiply(t *testing.T) {
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

func TestExprBinaryDivide(t *testing.T) {
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

func TestPerformRemainder(t *testing.T) {
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
