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
func newOperatorNode(eval *evaluator, token token, arity int) *operatorNode {
	text := token.contents()
	node := &operatorNode{exprNode{token.typ, text, eval}, arity, nil, nil, 0, false}
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
		return o.performRemainder()
	case "*":
		return o.performMultiply()
	case "/":
		return o.performDivide()
	case "+":
		return o.performPlus()
	case "-":
		return o.performMinus()
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

// performPlus performs the plus (+) unary and binary operators.
func (o *operatorNode) performPlus() (interface{}, *TclError) {
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
		if left == nil || right == nil {
			return nil, NewTclError(EOPERAND, "cannot operate on nil")
		}
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
	panic("unreachable")
}

// performMinus performs the minus (-) unary and binary operators.
func (o *operatorNode) performMinus() (interface{}, *TclError) {
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
		if left == nil || right == nil {
			return nil, NewTclError(EOPERAND, "cannot operate on nil")
		}
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
	panic("unreachable")
}

// performMultiply performs the multiplication (*) binary operator.
func (o *operatorNode) performMultiply() (interface{}, *TclError) {
	// evaluate both children, return product
	left, err := o.left.evaluate()
	if err != nil {
		return nil, err
	}
	right, err := o.right.evaluate()
	if err != nil {
		return nil, err
	}
	if left == nil || right == nil {
		return nil, NewTclError(EOPERAND, "cannot operate on nil")
	}
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

// performDivide performs the division (/) binary operator.
func (o *operatorNode) performDivide() (interface{}, *TclError) {
	// evaluate both children, return quotient
	left, err := o.left.evaluate()
	if err != nil {
		return nil, err
	}
	right, err := o.right.evaluate()
	if err != nil {
		return nil, err
	}
	if left == nil || right == nil {
		return nil, NewTclError(EOPERAND, "cannot operate on nil")
	}
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

// performRemainder performs the remainder (/) binary operator.
func (o *operatorNode) performRemainder() (interface{}, *TclError) {
	// evaluate both children, return remainder
	left, err := o.left.evaluate()
	if err != nil {
		return nil, err
	}
	right, err := o.right.evaluate()
	if err != nil {
		return nil, err
	}
	if left == nil || right == nil {
		return nil, NewTclError(EOPERAND, "cannot operate on nil")
	}
	li, li_ok := left.(int64)
	ri, ri_ok := right.(int64)
	if li_ok && ri_ok {
		return li % ri, nil
	}
	return nil, NewTclError(EOPERAND, "cannot operate on non-integer values")
}
