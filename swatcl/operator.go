//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"fmt"
)

type OperatorNode interface {
	ExprNode
	// getArity returns 1 for unary operators and 2 for binary.
	getArity() int
	// getPrecedence returns the precedence of the operator (lower number,
	// higher precedence.)
	getPrecedence() int
	// isSentinel returns true if this is not a real operator, such as a parenthesis.
	isSentinel() bool
	// setLeft assigns the given node as the "left" child of this operator.
	setLeft(left ExprNode)
	// setRight assigns the given node as the "right" child of this operator.
	setRight(right ExprNode)
}

// operatoreNode represents an operator (e.g. +, &) in Tcl. It has
// an arity (e.g. binary, unary) and a precedence.
type operatorNode struct {
	exprNode
	arity      int      // 1 for unary, 2 for binary
	left       ExprNode // left child node
	right      ExprNode // right child node
	precedence int      // operator precedence (with 1 being the highest)
	sentinel   bool     // true if this is a sentinel node (e.g. left parenthesis)
}

// newOperatorNode constructs an operator node based on the given attributes.
func newOperatorNode(eval Evaluator, token token, arity int) *operatorNode {
	text := token.contents()
	node := &operatorNode{exprNode{token.typ, text, eval}, arity, nil, nil, 0, false}
	// determine operator precedence
	switch text {
	case "(", ")":
		node.precedence = 1
		node.sentinel = true
	case "~", "!": // bitwise/logical complement?
		node.precedence = 4
	case "**": // power??
		node.precedence = 5
	case "*", "/", "%": // multiplicative
		node.precedence = 6
	case "+", "-": // additive
		if arity == 1 {
			node.precedence = 4
		} else {
			node.precedence = 7
		}
	case "<<", ">>": // shift
		node.precedence = 8
	case "<", ">", "<=", ">=": // relational
		node.precedence = 9
	case "eq", "ne", "in", "ni":
		node.precedence = 10
	case "&": // bitwise AND
		node.precedence = 11
	case "^": // bitwise exclusive OR
		node.precedence = 12
	case "|": // bitwise inclusive OR
		node.precedence = 13
	case "&&": // logical AND
		node.precedence = 14
	case "||": // logical OR
		node.precedence = 15
	case "?": // ternary
		node.precedence = 16
	default:
		node.precedence = -1
	}
	return node
}

// String returns a programmer friendly version of the operator.
func (o *operatorNode) String() string {
	return fmt.Sprintf("'%s' binary: %t, precedence: %d",
		o.text, o.arity == 2, o.precedence)
}

// getArity returns 1 for unary operators and 2 for binary.
func (o *operatorNode) getArity() int {
	return o.arity
}

// getPrecedence returns the precedence of the operator (lower number, higher
// precedence.)
func (o *operatorNode) getPrecedence() int {
	return o.precedence
}

// isSentinel returns true if this is not a real operator, such as a parenthesis.
func (o *operatorNode) isSentinel() bool {
	return o.sentinel
}

// setLeft assigns the given node as the "left" child of this operator.
func (o *operatorNode) setLeft(left ExprNode) {
	o.left = left
}

// setRight assigns the given node as the "right" child of this operator.
func (o *operatorNode) setRight(right ExprNode) {
	o.right = right
}

// evaluate will evaluate the left and right children (or just the right child
// for unary operators) and invoke the operation, returning the result.
func (o *operatorNode) evaluate() (interface{}, returnCode, *TclError) {
	var result interface{}
	if o.arity == 1 {
		right, code, err := o.right.evaluate()
		if err != nil {
			return nil, code, err
		}
		if right == nil {
			return nil, returnError, NewTclError(EOPERAND, o.text+" cannot operate on nil")
		}
		if o.text == "+" {
			result, err = operatorUnaryPlus(right)
		} else if o.text == "-" {
			result, err = operatorUnaryMinus(right)
		} else {
			return nil, returnError, NewTclError(EOPERATOR, "unsupported unary operator: "+o.text)
		}
		if err != nil {
			return nil, returnError, err
		}
		return result, returnOk, nil
	}
	left, code, err := o.left.evaluate()
	if err != nil {
		return nil, code, err
	}
	right, code, err := o.right.evaluate()
	if err != nil {
		return nil, code, err
	}
	if left == nil || right == nil {
		return nil, returnError, NewTclError(EOPERAND, o.text+" cannot operate on nil")
	}
	// TODO: based on the operator token, call the appropriate function with the operand(s)
	// TODO: see http://www.tcl.tk/man/tcl8.5/TclCmd/expr.htm for details
	switch o.text {
	// case "~", "!":
	// case "**":
	case "%":
		result, err = operatorRemainder(left, right)
	case "*":
		result, err = operatorMultiply(left, right)
	case "/":
		result, err = operatorDivide(left, right)
	case "+":
		result, err = operatorBinaryPlus(left, right)
	case "-":
		result, err = operatorBinaryMinus(left, right)
	// case "<<", ">>":
	// case "<", ">", "<=", ">=":
	// case "in", "ni":
	case "eq":
		result, err = operatorStringEqual(left, right)
	case "ne":
		result, err = operatorStringNotEqual(left, right)
	// case "&":
	// case "^":
	// case "|":
	// case "&&":
	// case "||":
	// case "?":
	default:
		panic(fmt.Sprintf("unknown operator '%s'", o.text))
	}
	if err != nil {
		return nil, returnError, err
	}
	return result, returnOk, nil
}

// operatorUnaryPlus performs the plus (+) unary operation.
func operatorUnaryPlus(val interface{}) (interface{}, *TclError) {
	switch n := val.(type) {
	case nil:
		return nil, NewTclError(EOPERAND, "cannot operate on nil")
	case int64:
		return 0 + n, nil
	case float64:
		return 0 + n, nil
	default:
		return nil, NewTclError(EOPERAND,
			fmt.Sprintf("unsupported operand type '%T' for '%s'", n, val))
	}
	panic("unreachable code")
}

// operatorBinaryPlus performs the plus (+) binary operation.
func operatorBinaryPlus(left, right interface{}) (interface{}, *TclError) {
	lf, lf_ok := left.(float64)
	rf, rf_ok := right.(float64)
	li, li_ok := left.(int64)
	ri, ri_ok := right.(int64)
	if lf_ok && rf_ok {
		return lf + rf, nil
	} else if lf_ok && ri_ok {
		return lf + float64(ri), nil
	} else if li_ok && rf_ok {
		return float64(li) + rf, nil
	} else if li_ok && ri_ok {
		return li + ri, nil
	}
	return nil, NewTclError(EOPERAND, "cannot operate on non-numeric values")
}

// operatorUnaryMinus performs the minus (-) unary operation.
func operatorUnaryMinus(val interface{}) (interface{}, *TclError) {
	switch n := val.(type) {
	case nil:
		return nil, NewTclError(EOPERAND, "cannot operate on nil")
	case int64:
		return 0 - n, nil
	case float64:
		return 0 - n, nil
	default:
		return nil, NewTclError(EOPERAND,
			fmt.Sprintf("unsupported operand type '%T' for '%s'", n, val))
	}
	panic("unreachable code")
}

// operatorBinaryMinus performs the minus (-) binary operation.
func operatorBinaryMinus(left, right interface{}) (interface{}, *TclError) {
	lf, lf_ok := left.(float64)
	rf, rf_ok := right.(float64)
	li, li_ok := left.(int64)
	ri, ri_ok := right.(int64)
	if lf_ok && rf_ok {
		return lf - rf, nil
	} else if lf_ok && ri_ok {
		return lf - float64(ri), nil
	} else if li_ok && rf_ok {
		return float64(li) - rf, nil
	} else if li_ok && ri_ok {
		return li - ri, nil
	}
	return nil, NewTclError(EOPERAND, "cannot operate on non-numeric values")
}

// operatorMultiply performs the multiplication (*) binary operator.
func operatorMultiply(left, right interface{}) (interface{}, *TclError) {
	lf, lf_ok := left.(float64)
	rf, rf_ok := right.(float64)
	li, li_ok := left.(int64)
	ri, ri_ok := right.(int64)
	if lf_ok && rf_ok {
		return lf * rf, nil
	} else if lf_ok && ri_ok {
		return lf * float64(ri), nil
	} else if li_ok && rf_ok {
		return float64(li) * rf, nil
	} else if li_ok && ri_ok {
		return li * ri, nil
	}
	return nil, NewTclError(EOPERAND, "cannot operate on non-numeric values")
}

// operatorDivide performs the division (/) binary operator.
func operatorDivide(left, right interface{}) (interface{}, *TclError) {
	lf, lf_ok := left.(float64)
	rf, rf_ok := right.(float64)
	li, li_ok := left.(int64)
	ri, ri_ok := right.(int64)
	if lf_ok && rf_ok {
		return lf / rf, nil
	} else if lf_ok && ri_ok {
		return lf / float64(ri), nil
	} else if li_ok && rf_ok {
		return float64(li) / rf, nil
	} else if li_ok && ri_ok {
		return li / ri, nil
	}
	return nil, NewTclError(EOPERAND, "cannot operate on non-numeric values")
}

// operatorRemainder performs the remainder (%) binary operator.
func operatorRemainder(left, right interface{}) (interface{}, *TclError) {
	li, li_ok := left.(int64)
	ri, ri_ok := right.(int64)
	if li_ok && ri_ok {
		return li % ri, nil
	}
	return nil, NewTclError(EOPERAND, "cannot operate on non-integer values")
}

// operatorStringEqual converts both arugments to strings and compares them,
// returning 1 if they are equal and 0 otherwise.
func operatorStringEqual(left, right interface{}) (interface{}, *TclError) {
	ls := fmt.Sprint(left)
	rs := fmt.Sprint(right)
	if ls == rs {
		return 1, nil
	}
	return 0, nil
}

// operatorStringNotEqual converts both arugments to strings and compares
// them, returning 1 if they are _not_ equal and 0 otherwise.
func operatorStringNotEqual(left, right interface{}) (interface{}, *TclError) {
	ls := fmt.Sprint(left)
	rs := fmt.Sprint(right)
	if ls == rs {
		return 0, nil
	}
	return 1, nil
}
