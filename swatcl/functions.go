//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"code.google.com/p/goswat/container/vector"
	"math"
	"math/rand"
	"strings"
)

// FunctionNode is similar to an operator except that it has zero or more
// arguments, given in a particular order.
type FunctionNode interface {
	ExprNode
	// PushArgument inserts the argument at the beginning of the list,
	// useful for the evaluator which is pulling the arguments off of a
	// stack in reverse order.
	PushArgument(a interface{})
}

type functionNode struct {
	exprNode
	arguments vector.Vector // function arguments
}

// newFunctionNode constructs a new function node. Any arguments may be added
// to the node as they are encountered by the interpreter.
func newFunctionNode(eval Evaluator, token token) FunctionNode {
	text := token.contents()
	return &functionNode{exprNode{token.typ, text, eval}, nil}
}

func (f *functionNode) PushArgument(a interface{}) {
	f.arguments.Insert(0, a)
}

// functionTable maps Tcl expression functions to implementations.
var functionTable = make(map[string]func([]interface{}) TclResult)

// populateFunctionTable adds the supported math functions to the table for
// ease of converting from a string (e.g. "abs") to an implementation.
func populateFunctionTable() {
	functionTable["abs"] = tclAbs
	functionTable["bool"] = tclBool
	functionTable["ceil"] = tclCeil
	functionTable["double"] = tclDouble
	functionTable["exp"] = tclExp
	functionTable["floor"] = tclFloor
	functionTable["fmod"] = tclFmod
	functionTable["log"] = tclLog
	functionTable["log10"] = tclLog10
	functionTable["max"] = tclMax
	functionTable["min"] = tclMin
	functionTable["pow"] = tclPow
	functionTable["rand"] = tclRand
	functionTable["round"] = tclRound
	functionTable["sqrt"] = tclSqrt
	functionTable["srand"] = tclSrand
}

// TODO: expose the math functions above in the tcl::matchfunc namespace as commands

// evaluate will evalute the function arguments and then invoke the function
// using those arguments, returning the result.
func (f *functionNode) evaluate() TclResult {
	if len(functionTable) == 0 {
		populateFunctionTable()
	}
	fn := functionTable[f.text]
	if fn == nil {
		return newTclResultErrorf(ECOMMAND, "unsupported function '%s'", f.text)
	}
	// evaluate the arguments and invoke the function
	var result TclResult
	args := make([]interface{}, 0, len(f.arguments))
	for _, a := range f.arguments {
		en, ok := a.(ExprNode)
		if !ok {
			return newTclResultErrorf(EARGUMENT, "function argument wrong type: '%v' (%T)", a, a)
		}
		result = en.evaluate()
		if !result.Ok() {
			return result
		}
		// try to coerce the values into numbers
		val, err := coerceNumber(result.Result())
		if err != nil {
			return newTclResultError(ESYNTAX, err.Error())
		}
		args = append(args, val)
	}
	return fn(args)
}

// getText returns the name of the function.
func (f *functionNode) getText() string {
	return f.text
}

// getArity always returns zero since functions take variable arguments.
func (f *functionNode) getArity() int {
	return 0
}

// getPrecedence returns 1 for functions.
func (f *functionNode) getPrecedence() int {
	return 17
}

// isSentinel is always false for functions.
func (f *functionNode) isSentinel() bool {
	return false
}

// setLeft does nothing for functions.
func (f *functionNode) setLeft(left ExprNode) {
	// do nothing
}

// setRight does nothing for functions.
func (f *functionNode) setRight(right ExprNode) {
	// do nothing
}

// tclAbs implements the tcl::mathfunc::abs command.
func tclAbs(args []interface{}) TclResult {
	if len(args) != 1 {
		return newTclResultError(EARGUMENT, "abs() takes exactly one argument")
	}
	var result interface{} = nil
	fl, flok := args[0].(float64)
	if flok {
		result = math.Abs(fl)
	}
	in, inok := args[0].(int64)
	if inok {
		result = int64(math.Abs(float64(in)))
	}
	if result != nil {
		return newOperatorResult(result)
	}
	return newTclResultError(EARGUMENT, "abs() takes only ints and floats")
}

// tclBool implements the tcl::mathfunc::bool command.
func tclBool(args []interface{}) TclResult {
	if len(args) != 1 {
		return newTclResultError(EARGUMENT, "bool() takes exactly one argument")
	}
	fl, ok := args[0].(float64)
	if ok {
		return newOperatorResult(fl != 0)
	}
	in, ok := args[0].(int64)
	if ok {
		return newOperatorResult(in != 0)
	}
	str, ok := args[0].(string)
	if ok {
		str = strings.ToLower(str)
		switch str {
		case "0", "false", "no", "off":
			return newOperatorResult(false)
		case "1", "true", "yes", "on":
			return newOperatorResult(true)
		default:
			return newTclResultError(EARGUMENT, "expected 'string is boolean' value")
		}
	}
	return newTclResultError(EARGUMENT, "bool() takes only ints and floats")
}

// tclCeil implements the tcl::mathfunc::ceil command.
func tclCeil(args []interface{}) TclResult {
	if len(args) != 1 {
		return newTclResultError(EARGUMENT, "ceil() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return newOperatorResult(math.Ceil(fl))
	}
	in, inok := args[0].(int64)
	if inok {
		return newOperatorResult(math.Ceil(float64(in)))
	}
	return newTclResultError(EARGUMENT, "ceil() takes only ints and floats")
}

// tclDouble implements the tcl::mathfunc::double command.
func tclDouble(args []interface{}) TclResult {
	if len(args) != 1 {
		return newTclResultError(EARGUMENT, "double() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return newOperatorResult(fl)
	}
	in, inok := args[0].(int64)
	if inok {
		return newOperatorResult(float64(in))
	}
	return newTclResultError(EARGUMENT, "double() takes only ints and floats")
}

// tclExp implements the tcl::mathfunc::exp command.
func tclExp(args []interface{}) TclResult {
	if len(args) != 1 {
		return newTclResultError(EARGUMENT, "exp() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return newOperatorResult(math.Exp(fl))
	}
	in, inok := args[0].(int64)
	if inok {
		return newOperatorResult(math.Exp(float64(in)))
	}
	return newTclResultError(EARGUMENT, "exp() takes only ints and floats")
}

// tclFloor implements the tcl::mathfunc::floor command.
func tclFloor(args []interface{}) TclResult {
	if len(args) != 1 {
		return newTclResultError(EARGUMENT, "floor() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return newOperatorResult(math.Floor(fl))
	}
	in, inok := args[0].(int64)
	if inok {
		return newOperatorResult(math.Floor(float64(in)))
	}
	return newTclResultError(EARGUMENT, "floor() takes only ints and floats")
}

// tclFmod implements the tcl::mathfunc::fmod command.
func tclFmod(args []interface{}) TclResult {
	if len(args) != 2 {
		return newTclResultError(EARGUMENT, "fmod() takes exactly two arguments")
	}
	flx, flok := args[0].(float64)
	if !flok {
		in, inok := args[0].(int64)
		if inok {
			flx = float64(in)
		} else {
			return newTclResultError(EARGUMENT, "fmod() takes only ints and floats")
		}
	}
	fly, flok := args[1].(float64)
	if !flok {
		in, inok := args[1].(int64)
		if inok {
			fly = float64(in)
		} else {
			return newTclResultError(EARGUMENT, "fmod() takes only ints and floats")
		}
	}
	return newOperatorResult(math.Mod(flx, fly))
}

// tclLog implements the tcl::mathfunc::log command.
func tclLog(args []interface{}) TclResult {
	if len(args) != 1 {
		return newTclResultError(EARGUMENT, "log() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return newOperatorResult(math.Log(fl))
	}
	in, inok := args[0].(int64)
	if inok {
		return newOperatorResult(math.Log(float64(in)))
	}
	return newTclResultError(EARGUMENT, "log() takes only ints and floats")
}

// tclLog10 implements the tcl::mathfunc::log10 command.
func tclLog10(args []interface{}) TclResult {
	if len(args) != 1 {
		return newTclResultError(EARGUMENT, "log10() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return newOperatorResult(math.Log10(fl))
	}
	in, inok := args[0].(int64)
	if inok {
		return newOperatorResult(math.Log10(float64(in)))
	}
	return newTclResultError(EARGUMENT, "log10() takes only ints and floats")
}

// tclMax implements the tcl::mathfunc::max command.
func tclMax(args []interface{}) TclResult {
	if len(args) < 1 {
		return newTclResultError(EARGUMENT, "max() takes at least one argument")
	}
	// scan to check that all arguments are numbers
	// also see if they are all ints or not
	all_ints := true
	for _, n := range args {
		if _, flok := n.(float64); flok {
			all_ints = false
		} else if _, inok := n.(int64); !inok {
			return newTclResultError(EARGUMENT, "max() takes only ints and floats")
		}
	}
	if all_ints {
		var max int64 = -9223372036854775808
		for _, n := range args {
			in, _ := n.(int64)
			if in > max {
				max = in
			}
		}
		return newOperatorResult(max)
	} else {
		var max float64 = math.Inf(-1)
		for _, n := range args {
			fl, ok := n.(float64)
			if !ok {
				if in, ok := n.(int64); ok {
					fl = float64(in)
				}
			}
			if fl > max {
				max = fl
			}
		}
		return newOperatorResult(max)
	}
	panic("unreachable code")
}

// tclMin implements the tcl::mathfunc::min command.
func tclMin(args []interface{}) TclResult {
	if len(args) < 1 {
		return newTclResultError(EARGUMENT, "min() takes at least one argument")
	}
	// scan to check that all arguments are numbers
	// also see if they are all ints or not
	all_ints := true
	for _, n := range args {
		if _, flok := n.(float64); flok {
			all_ints = false
		} else if _, inok := n.(int64); !inok {
			return newTclResultError(EARGUMENT, "min() takes only ints and floats")
		}
	}
	if all_ints {
		var min int64 = 9223372036854775807
		for _, n := range args {
			in, _ := n.(int64)
			if in < min {
				min = in
			}
		}
		return newOperatorResult(min)
	} else {
		var min float64 = math.Inf(1)
		for _, n := range args {
			fl, ok := n.(float64)
			if !ok {
				if in, ok := n.(int64); ok {
					fl = float64(in)
				}
			}
			if fl < min {
				min = fl
			}
		}
		return newOperatorResult(min)
	}
	panic("unreachable code")
}

// tclPow implements the tcl::mathfunc::pow command.
func tclPow(args []interface{}) TclResult {
	if len(args) != 2 {
		return newTclResultError(EARGUMENT, "pow() takes exactly two arguments")
	}
	flx, flok := args[0].(float64)
	if !flok {
		in, inok := args[0].(int64)
		if inok {
			flx = float64(in)
		} else {
			return newTclResultError(EARGUMENT, "pow() takes only ints and floats")
		}
	}
	fly, flok := args[1].(float64)
	if !flok {
		in, inok := args[1].(int64)
		if inok {
			fly = float64(in)
		} else {
			return newTclResultError(EARGUMENT, "pow() takes only ints and floats")
		}
	}
	return newOperatorResult(math.Pow(flx, fly))
}

// tclRand implements the tcl::mathfunc::rand command.
func tclRand(args []interface{}) TclResult {
	if len(args) != 0 {
		return newTclResultError(EARGUMENT, "rand() takes no arguments")
	}
	return newOperatorResult(rand.Float64())
}

// mathRound rounds the given floating point number to the nearest integer,
// rounding to even if the fractional part is equal to .5, as required by
// Scheme R5RS (also happens to be the IEEE 754 recommended default).
func mathRound(num float64) (int64, error) {
	in, fr := math.Modf(num)
	if math.IsNaN(fr) {
		return 0, NumberOutOfRange
	}
	fr = math.Abs(fr)
	ini := int64(in)
	if fr < 0.5 {
		return ini, nil
	}
	if fr > 0.5 {
		if ini > 0 {
			return ini + 1, nil
		}
		return ini - 1, nil
	}
	if ini&1 == 0 {
		return ini, nil
	}
	if ini > 0 {
		return ini + 1, nil
	}
	return ini - 1, nil
}

// tclRound implements the tcl::mathfunc::round command.
func tclRound(args []interface{}) TclResult {
	if len(args) != 1 {
		return newTclResultError(EARGUMENT, "round() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		result, err := mathRound(fl)
		if err != nil {
			return newTclResultError(EARGUMENT, err.Error())
		}
		return newOperatorResult(result)
	}
	in, inok := args[0].(int64)
	if inok {
		return newOperatorResult(in)
	}
	return newTclResultError(EARGUMENT, "round() takes only ints and floats")
}

// tclSqrt implements the tcl::mathfunc::sqrt command.
func tclSqrt(args []interface{}) TclResult {
	if len(args) != 1 {
		return newTclResultError(EARGUMENT, "sqrt() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return newOperatorResult(math.Sqrt(fl))
	}
	in, inok := args[0].(int64)
	if inok {
		return newOperatorResult(math.Sqrt(float64(in)))
	}
	return newTclResultError(EARGUMENT, "sqrt() takes only ints and floats")
}

// tclSrand implements the tcl::mathfunc::srand command.
func tclSrand(args []interface{}) TclResult {
	if len(args) != 1 {
		return newTclResultError(EARGUMENT, "srand() takes exactly one argument")
	}
	in, inok := args[0].(int64)
	if inok {
		rand.Seed(in)
		return newOperatorResult(rand.Float64())
	}
	return newTclResultError(EARGUMENT, "srand() takes only integers")
}

// TODO: support following math functions
// acos, asin, atan, atan2, cos, cosh, entier, hypot, int
// isqrt, sin, sinh, tan, tanh, wide
