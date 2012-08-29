//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"code.google.com/p/goswat/container/vector"
	"fmt"
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
func newFunctionNode(eval *evaluator, token token) FunctionNode {
	text := token.contents()
	return &functionNode{exprNode{token.typ, text, eval}, nil}
}

func (f *functionNode) PushArgument(a interface{}) {
	f.arguments.Insert(0, a)
}

// functionTable maps Tcl expression functions to implementations.
var functionTable = make(map[string]func([]interface{}) (interface{}, *TclError))

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
	functionTable["sqrt"] = tclSqrt
	functionTable["srand"] = tclSrand
}

// TODO: expose the math functions above in the tcl::matchfunc namespace as commands

// evaluate will evalute the function arguments and then invoke the function
// using those arguments, returning the result.
func (f *functionNode) evaluate() (interface{}, *TclError) {
	if len(functionTable) == 0 {
		populateFunctionTable()
	}
	fn := functionTable[f.text]
	if fn == nil {
		return nil, NewTclError(ENOFUNC, "unsupported function "+f.text)
	}
	// evaluate the arguments and invoke the function
	args := make([]interface{}, 0, len(f.arguments))
	for _, a := range f.arguments {
		en, ok := a.(ExprNode)
		if !ok {
			msg := fmt.Sprintf("function argument wrong type: '%v' (%T)", a, a)
			return nil, NewTclError(EILLARG, msg)
		}
		val, err := en.evaluate()
		if err != nil {
			return nil, err
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
func tclAbs(args []interface{}) (interface{}, *TclError) {
	if len(args) != 1 {
		return nil, NewTclError(EILLARG, "abs() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return math.Abs(fl), nil
	}
	in, inok := args[0].(int64)
	if inok {
		return int64(math.Abs(float64(in))), nil
	}
	return nil, NewTclError(EILLARG, "abs() takes only ints and floats")
}

// tclBool implements the tcl::mathfunc::bool command.
func tclBool(args []interface{}) (interface{}, *TclError) {
	if len(args) != 1 {
		return nil, NewTclError(EILLARG, "bool() takes exactly one argument")
	}
	fl, ok := args[0].(float64)
	if ok {
		return fl != 0, nil
	}
	in, ok := args[0].(int64)
	if ok {
		return in != 0, nil
	}
	str, ok := args[0].(string)
	if ok {
		str = strings.ToLower(str)
		switch str {
		case "0", "false", "no", "off":
			return false, nil
		case "1", "true", "yes", "on":
			return true, nil
		default:
			return nil, NewTclError(EILLARG, "expected 'string is boolean' value")
		}
	}
	return nil, NewTclError(EILLARG, "bool() takes only ints and floats")
}

// tclCeil implements the tcl::mathfunc::ceil command.
func tclCeil(args []interface{}) (interface{}, *TclError) {
	if len(args) != 1 {
		return nil, NewTclError(EILLARG, "ceil() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return math.Ceil(fl), nil
	}
	in, inok := args[0].(int64)
	if inok {
		return math.Ceil(float64(in)), nil
	}
	return nil, NewTclError(EILLARG, "ceil() takes only ints and floats")
}

// tclDouble implements the tcl::mathfunc::double command.
func tclDouble(args []interface{}) (interface{}, *TclError) {
	if len(args) != 1 {
		return nil, NewTclError(EILLARG, "double() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return fl, nil
	}
	in, inok := args[0].(int64)
	if inok {
		return float64(in), nil
	}
	return nil, NewTclError(EILLARG, "double() takes only ints and floats")
}

// tclExp implements the tcl::mathfunc::exp command.
func tclExp(args []interface{}) (interface{}, *TclError) {
	if len(args) != 1 {
		return nil, NewTclError(EILLARG, "exp() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return math.Exp(fl), nil
	}
	in, inok := args[0].(int64)
	if inok {
		return math.Exp(float64(in)), nil
	}
	return nil, NewTclError(EILLARG, "exp() takes only ints and floats")
}

// tclFloor implements the tcl::mathfunc::floor command.
func tclFloor(args []interface{}) (interface{}, *TclError) {
	if len(args) != 1 {
		return nil, NewTclError(EILLARG, "floor() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return math.Floor(fl), nil
	}
	in, inok := args[0].(int64)
	if inok {
		return math.Floor(float64(in)), nil
	}
	return nil, NewTclError(EILLARG, "floor() takes only ints and floats")
}

// tclFmod implements the tcl::mathfunc::fmod command.
func tclFmod(args []interface{}) (interface{}, *TclError) {
	if len(args) != 2 {
		return nil, NewTclError(EILLARG, "fmod() takes exactly two arguments")
	}
	flx, flok := args[0].(float64)
	if !flok {
		in, inok := args[0].(int64)
		if inok {
			flx = float64(in)
		} else {
			return nil, NewTclError(EILLARG, "fmod() takes only ints and floats")
		}
	}
	fly, flok := args[1].(float64)
	if !flok {
		in, inok := args[1].(int64)
		if inok {
			fly = float64(in)
		} else {
			return nil, NewTclError(EILLARG, "fmod() takes only ints and floats")
		}
	}
	return math.Mod(flx, fly), nil
}

// tclLog implements the tcl::mathfunc::log command.
func tclLog(args []interface{}) (interface{}, *TclError) {
	if len(args) != 1 {
		return nil, NewTclError(EILLARG, "log() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return math.Log(fl), nil
	}
	in, inok := args[0].(int64)
	if inok {
		return math.Log(float64(in)), nil
	}
	return nil, NewTclError(EILLARG, "log() takes only ints and floats")
}

// tclLog10 implements the tcl::mathfunc::log10 command.
func tclLog10(args []interface{}) (interface{}, *TclError) {
	if len(args) != 1 {
		return nil, NewTclError(EILLARG, "log10() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return math.Log10(fl), nil
	}
	in, inok := args[0].(int64)
	if inok {
		return math.Log10(float64(in)), nil
	}
	return nil, NewTclError(EILLARG, "log10() takes only ints and floats")
}

// tclMax implements the tcl::mathfunc::max command.
func tclMax(args []interface{}) (interface{}, *TclError) {
	if len(args) < 1 {
		return nil, NewTclError(EILLARG, "max() takes at least one argument")
	}
	// scan to check that all arguments are numbers
	// also see if they are all ints or not
	all_ints := true
	for _, n := range args {
		if _, flok := n.(float64); flok {
			all_ints = false
		} else if _, inok := n.(int64); !inok {
			return nil, NewTclError(EILLARG, "max() takes only ints and floats")
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
		return max, nil
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
		return max, nil
	}
	panic("unreachable code")
}

// tclMin implements the tcl::mathfunc::min command.
func tclMin(args []interface{}) (interface{}, *TclError) {
	if len(args) < 1 {
		return nil, NewTclError(EILLARG, "min() takes at least one argument")
	}
	// scan to check that all arguments are numbers
	// also see if they are all ints or not
	all_ints := true
	for _, n := range args {
		if _, flok := n.(float64); flok {
			all_ints = false
		} else if _, inok := n.(int64); !inok {
			return nil, NewTclError(EILLARG, "min() takes only ints and floats")
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
		return min, nil
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
		return min, nil
	}
	panic("unreachable code")
}

// tclPow implements the tcl::mathfunc::pow command.
func tclPow(args []interface{}) (interface{}, *TclError) {
	if len(args) != 2 {
		return nil, NewTclError(EILLARG, "pow() takes exactly two arguments")
	}
	flx, flok := args[0].(float64)
	if !flok {
		in, inok := args[0].(int64)
		if inok {
			flx = float64(in)
		} else {
			return nil, NewTclError(EILLARG, "pow() takes only ints and floats")
		}
	}
	fly, flok := args[1].(float64)
	if !flok {
		in, inok := args[1].(int64)
		if inok {
			fly = float64(in)
		} else {
			return nil, NewTclError(EILLARG, "pow() takes only ints and floats")
		}
	}
	return math.Pow(flx, fly), nil
}

// tclRand implements the tcl::mathfunc::rand command.
func tclRand(args []interface{}) (interface{}, *TclError) {
	if len(args) != 0 {
		return nil, NewTclError(EILLARG, "rand() takes no arguments")
	}
	return rand.Float64(), nil
}

// tclSqrt implements the tcl::mathfunc::sqrt command.
func tclSqrt(args []interface{}) (interface{}, *TclError) {
	if len(args) != 1 {
		return nil, NewTclError(EILLARG, "sqrt() takes exactly one argument")
	}
	fl, flok := args[0].(float64)
	if flok {
		return math.Sqrt(fl), nil
	}
	in, inok := args[0].(int64)
	if inok {
		return math.Sqrt(float64(in)), nil
	}
	return nil, NewTclError(EILLARG, "sqrt() takes only ints and floats")
}

// tclSrand implements the tcl::mathfunc::srand command.
func tclSrand(args []interface{}) (interface{}, *TclError) {
	if len(args) != 1 {
		return nil, NewTclError(EILLARG, "srand() takes exactly one argument")
	}
	in, inok := args[0].(int64)
	if inok {
		rand.Seed(in)
		return rand.Float64(), nil
	}
	return nil, NewTclError(EILLARG, "srand() takes only integers")
}

// TODO // tclRound implements the tcl::mathfunc::round command.
// func tclRound(args []interface{}) (interface{}, *TclError) {
// 	if len(args) != 1 {
// 		return nil, NewTclError(EILLARG, "round() takes exactly one argument")
// 	}
// 	fl, flok := args[0].(float64)
// 	if flok {
// 		return math.Round(fl), nil
// 	}
// 	in, inok := args[0].(int64)
// 	if inok {
// 		return in, nil
// 	}
// 	return nil, NewTclError(EILLARG, "round() takes only ints and floats")
// }

// TODO: support following math functions
// acos, asin, atan, atan2, cos, cosh, entier, hypot, int
// isqrt, sin, sinh, tan, tanh, wide
