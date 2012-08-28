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
}

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
	return 1
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

// tclAbs implements the abs() function in Tcl expressions. It always
// returns a floating point result.
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
		return math.Abs(float64(in)), nil
	}
	return nil, NewTclError(EILLARG, "abs() takes only ints and floats")
}

// TODO: support following math functions
// acos        asin        atan
// atan2       bool        ceil        cos
// cosh        double      entier      exp
// floor       fmod        hypot       int
// isqrt       log         log10       max
// min         pow         rand        round
// sin         sinh        sqrt        srand
// tan         tanh        wide
// (where double, int, wide, entier are type conversions)
