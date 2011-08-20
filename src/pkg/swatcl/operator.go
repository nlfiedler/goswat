//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"fmt"
)

// operatoreNode represents an operator (e.g. +, &) in Tcl. It has
// an arity (e.g. binary, unary) and a precedence.
type operatorNode struct {
	exprNode
	arity      int       // 1 for unary, 2 for binary
	left       *exprNode // left child node
	right      *exprNode // right child node
	precedence int       // operator precedence (with 1 being the highest)
	sentinel   bool      // true if this is a sentinel node (e.g. left parenthesis)
}

// TODO: left paren is an operator node with precedence of 1 and a sentinel flag

// newOperatorNode constructs an operator node based on the given attributes.
func newOperatorNode(eval *evaluator, token parserToken, text string, arity int) *operatorNode {
	node := &operatorNode{exprNode{token, text, eval}, arity, nil, nil, 0, false}
	// determine operator precedence
	switch text {
	case "~", "!":
		node.precedence = 13
	case "**":
		node.precedence = 12
	case "*", "/", "%":
		node.precedence = 11
	case "+", "-":
		if arity == 1 {
			node.precedence = 13
		} else {
			node.precedence = 10
		}
	case "<<", ">>":
		node.precedence = 9
	case "<", ">", "<=", ">=":
		node.precedence = 8
	case "eq", "ne", "in", "ni":
		node.precedence = 7
	case "&":
		node.precedence = 6
	case "^":
		node.precedence = 5
	case "|":
		node.precedence = 4
	case "&&":
		node.precedence = 3
	case "||":
		node.precedence = 2
	case "?":
		node.precedence = 1
	default:
		node.precedence = -1
	}
	return node
}

func (o *operatorNode) evaluate() (interface{}, *TclError) {
	// TODO: based on the operator token, call the appropriate function with the operand(s)
	switch o.text {
	// case "~", "!":
	// case "**":
	case "%":
		return o.handleRemainder()
	case "*":
		return o.handleMultiply()
	case "/":
		return o.handleDivide()
	case "+":
		return o.handlePlus()
	case "-":
		return o.handleMinus()
	// case "<<", ">>":
	// case "<", ">", "<=", ">=":
	// case "eq", "ne", "in", "ni":
	// case "&":
	// case "^":
	// case "|":
	// case "&&":
	// case "||":
	// case "?":
	default:
		panic(fmt.Sprintf("unknown operator '%s'", o.text))
	}
	return "", nil
}

// handlePlus performs the plus (+) unary and binary operators.
func (o *operatorNode) handlePlus() (interface{}, *TclError) {
	if o.arity == 1 {
		// evaluate right child and return 0 + value
		val, err := o.right.evaluate()
		if err != nil {
			return nil, err
		}
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
	} else {
		// evaluate both children, return sum
		left, err := o.left.evaluate()
		if err != nil {
			return nil, err
		}
		right, err := o.right.evaluate()
		if err != nil {
			return nil, err
		}
		return performAddition(left, right)
	}
	panic("unreachable")
}

// handleMinus performs the minus (-) unary and binary operators.
func (o *operatorNode) handleMinus() (interface{}, *TclError) {
	if o.arity == 1 {
		// evaluate right child and return 0 - value
		val, err := o.right.evaluate()
		if err != nil {
			return nil, err
		}
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
	} else {
		// evaluate both children, return difference
		left, err := o.left.evaluate()
		if err != nil {
			return nil, err
		}
		right, err := o.right.evaluate()
		if err != nil {
			return nil, err
		}
		return performSubtraction(left, right)
	}
	panic("unreachable")
}

// handleMultiply performs the multiplication (*) binary operator.
func (o *operatorNode) handleMultiply() (interface{}, *TclError) {
	// evaluate both children, return product
	left, err := o.left.evaluate()
	if err != nil {
		return nil, err
	}
	right, err := o.right.evaluate()
	if err != nil {
		return nil, err
	}
	return performMultiplication(left, right)
}

// handleDivide performs the division (/) binary operator.
func (o *operatorNode) handleDivide() (interface{}, *TclError) {
	// evaluate both children, return quotient
	left, err := o.left.evaluate()
	if err != nil {
		return nil, err
	}
	right, err := o.right.evaluate()
	if err != nil {
		return nil, err
	}
	return performDivision(left, right)
}

// handleRemainder performs the remainder (/) binary operator.
func (o *operatorNode) handleRemainder() (interface{}, *TclError) {
	// evaluate both children, return remainder
	left, err := o.left.evaluate()
	if err != nil {
		return nil, err
	}
	right, err := o.right.evaluate()
	if err != nil {
		return nil, err
	}
	return performRemainder(left, right)
}
