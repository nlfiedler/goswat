//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

//
// Expression parser and evaluator for Tcl expressions.
//
// Infix expressions can be parsed from a series of tokens using two
// stacks. One stack is used to hold parse trees under construction (the
// argument stack), the other to hold operators (and left parentheses,
// for matching purposes; the operator stack).
//
// As we read in each new token (from left to right), we either push the
// token (or a related tree) onto one of the stacks, or we reduce the
// stacks by combining an operator with some arguments. Along the way,
// it will be helpful to maintain a search state which tells us whether
// we should see an argument or operator next (the search state helps us
// to reject malformed expressions).
//

import (
	"code.google.com/p/goswat/container/vector"
	"fmt"
)

// searchState indicates what the expression evaluator is expecting
// to see next, typically either an argument or an operator. This is
// used to determine whether certain operators are binary or unary,
// such as + and - which can be both.
type searchState int

const (
	_              = iota
	searchArgument // expecting an argument
	searchOperator // expecting an operator
)

// ExprNode represents a node in the expression tree that can be evaluated
// to a final value.
type ExprNode interface {
	evaluate() (interface{}, returnCode, *TclError)
	getText() string
}

// exprNode contains the attributes of an expression node, whether it
// is a numeric or string literal, or a command or variable reference.
type exprNode struct {
	token tokenType
	text  string
	eval  Evaluator
}

// newExprNode creates a new expression node based on the token and text.
func newExprNode(eval Evaluator, token token) *exprNode {
	return &exprNode{token.typ, token.contents(), eval}
}

// getText returns the original text of the node.
func (n *exprNode) getText() string {
	return n.text
}

// String returns a string representation of the node.
func (n *exprNode) String() string {
	return n.text
}

// evaluate evaluates the expression node appropriately based on its type.
func (n *exprNode) evaluate() (interface{}, returnCode, *TclError) {
	switch n.token {
	case tokenVariable:
		str, err := n.eval.GetVariable(n.text)
		if err != nil {
			return "", returnError, err
		}
		return coerceNumber(str)
	case tokenCommand:
		return n.eval.Interpret(n.text)
	case tokenInteger:
		return atoi(n.text)
	case tokenFloat:
		return atof(n.text)
	case tokenString, tokenQuote:
		// perform basic string substitution (slash escapes)
		return evalString(n.text)
	case tokenBrace:
		// return the string as-is
		return n.text, returnOk, nil
	}
	return "", returnOk, nil
}

// Evaluator knows how to evaluate a Tcl expression.
type Evaluator interface {
	// Evaluate parses the expression, evaluates it, and returns the
	// result.
	Evaluate(expr string) (string, returnCode, *TclError)
	// Interpret passes the expression to the associated interpreter and
	// returns the result.
	Interpret(expr string) (string, returnCode, *TclError)
	// GetVariable retrieves a variable from the interpreter.
	GetVariable(name string) (string, *TclError)
}

// evaluator is an implementation of the Evaluator interface.
type evaluator struct {
	state     searchState   // looking for argument or operator?
	root      ExprNode      // root of the expression tree
	interp    Interpreter   // contains program state
	arguments vector.Vector // argument stack
	operators vector.Vector // operator stack
	funcCount int           // depth of function call nesting
}

// newEvaluator constructs an instance of Evaluator and associates it with
// the given interpreter, which allows for executing nested commands.
func newEvaluator(interp Interpreter) Evaluator {
	e := &evaluator{
		state:  searchArgument,
		interp: interp,
	}
	return e
}

// dumpStacks prints the contents of the operator and argument stacks to the
// console, useful for debugging the expression evaluator.
func (e *evaluator) dumpStacks() {
	fmt.Println("Argument stack...")
	if e.arguments.Len() == 0 {
		fmt.Println("(empty)")
	}
	for ii := e.arguments.Len() - 1; ii >= 0; ii-- {
		node, ok := e.arguments.At(ii).(ExprNode)
		if !ok {
			fmt.Printf("%d: not an ExprNode!", ii)
		}
		fmt.Printf("%d: %v (%T)\n", ii, node, node)
	}

	fmt.Println("\nOperator stack...")
	if e.operators.Len() == 0 {
		fmt.Println("(empty)")
	}
	for ii := e.operators.Len() - 1; ii >= 0; ii-- {
		node, ok := e.operators.At(ii).(ExprNode)
		if !ok {
			fmt.Printf("%d: not an ExprNode!", ii)
		}
		fmt.Printf("%d: %v (%T)\n", ii, node, node)
	}
}

// Reduce the operator stack by one. If the element at the top of the operator
// stack is a sentinel, no change is made.
func (e *evaluator) reduce() *TclError {
	// If there is a binary operator on top of the operator stack, there
	// should be two trees on top of the argument stack, both representing
	// expressions. Pop the operator and two trees off of the argument
	// stack, combining them into a single tree node, which is then pushed
	// back on the argument stack. Note that the trees on the argument
	// stack represent the right and left arguments, respectively.
	top, ok := e.operators.Pop().(OperatorNode)
	if !ok {
		return NewTclError(EBADSTATE, "node on operator stack is not an operator!")
	}
	if top.isSentinel() {
		// Cleverly do nothing and let the caller handle it.
	} else if top.getArity() == 2 {
		if e.arguments.Len() < 2 {
			e.dumpStacks()
			return NewTclError(EOPERAND, "operator requires two arguments")
		}
		arg2, ok := e.arguments.Pop().(ExprNode)
		if !ok {
			return NewTclError(EBADSTATE, "second argument is not an exprNode!")
		}
		arg1, ok := e.arguments.Pop().(ExprNode)
		if !ok {
			return NewTclError(EBADSTATE, "first argument is not an exprNode!")
		}
		top.setLeft(arg1)
		top.setRight(arg2)
		e.arguments.Push(top)
	} else if top.getArity() == 1 {
		if e.arguments.Len() < 1 {
			return NewTclError(EOPERAND, "operator requires one argument")
		}
		node := e.arguments.Pop()
		arg, ok := node.(ExprNode)
		if !ok {
			msg := fmt.Sprintf("expected exprNode for single argument, got %T", node)
			return NewTclError(EBADSTATE, msg)
		}
		top.setRight(arg)
		e.arguments.Push(top)
	} else {
		return NewTclError(EOPERATOR, "unknown operator "+top.getText())
	}
	return nil
}

// handleEOF reduces the operator stack, if there is anything on the
// stack, and then sets the resulting argument node to the root.
func (e *evaluator) handleEOF() *TclError {
	// If there is only one tree on the argument stack and the
	// operator stack is empty, return the single tree as the
	// result. If there are more trees and/or operators, reduce the
	// stacks as far as possible.
	count := 0
	for e.operators.Len() > 0 {
		top, ok := e.operators.Last().(OperatorNode)
		if !ok {
			return NewTclError(EBADSTATE, "node on operator stack is not an operatorNode!")
		}
		if top.getText() == "(" {
			return NewTclError(EPAREN, "unmatched left parenthesis")
		}
		if top.isSentinel() {
			return NewTclError(EBADEXPR, "sentinel operator encountered")
		}
		err := e.reduce()
		if err != nil {
			return err
		}
		if count++; count > 500 {
			return NewTclError(EBADSTATE, "operator stack too large")
		}
	}
	if e.arguments.Len() > 0 {
		topArg, ok := e.arguments.Pop().(ExprNode)
		if !ok {
			return NewTclError(EBADSTATE, "node on argument stack is not an ExprNode!")
		}
		if e.arguments.Len() == 0 && e.operators.Len() == 0 {
			e.root = topArg
		} else {
			return NewTclError(EBADSTATE, "argument stack is not empty")
		}
	}
	return nil
}

// handleCloseParen reduces the operator stack until it find the
// left parenthesis, signaling an error if this does not succeed.
func (e *evaluator) handleCloseParen() *TclError {
	if e.operators.Len() == 0 {
		return NewTclError(EPAREN, "unmatched right parenthesis")
	}
	// If there is a left parenthesis on the operator stack, we can
	// "cancel" the pair. If the operator stack contains some other
	// operator on top, reduce the stacks. This also covers the case
	// where the parentheses were used for grouping only.
	top, ok := e.operators.Last().(OperatorNode)
	for ok && !top.isSentinel() {
		if err := e.reduce(); err != nil {
			return err
		}
		if e.operators.Len() == 0 {
			return NewTclError(EPAREN, "unmatched right parenthesis")
		}
		top, ok = e.operators.Last().(OperatorNode)
	}
	if top != nil && top.getText() == "(" {
		// Remove the open parenthesis now that we're done reducing.
		e.operators.Pop()
	}
	if e.operators.Len() == 0 {
		// we're done
		return nil
	}
	// e.dumpStacks()
	// Now check for the function invocation case.
	if fun, ok := e.operators.Last().(FunctionNode); ok {
		// Take the function off of the operator stack, and then
		// remove nodes from the arguments stack until we see the
		// function node.
		e.operators.Pop()
		_, ok := e.arguments.Last().(FunctionNode)
		for !ok {
			arg := e.arguments.Pop()
			// Shove each new argument in the front of the list.
			fun.PushArgument(arg)
			_, ok = e.arguments.Last().(FunctionNode)
		}
		e.arguments.Pop()
		// Put the function invocation back on the argument stack
		// because it is treated as a value, not an operator.
		e.arguments.Push(fun)
		e.funcCount--
	}
	return nil
}

// handleComma attempts to reduce the operator stack with the assumption
// that the comma is being used to separate arguments to a function
// invocation.
func (e *evaluator) handleComma() *TclError {
	if e.funcCount == 0 {
		return NewTclError(EBADEXPR, "found comma outside function call")
	} else {
		// Reduce the operator stack to the left parenthesis.
		if e.operators.Len() < 2 {
			return NewTclError(EBADEXPR, "found comma outside function call")
		}
		top, ok := e.operators.Last().(OperatorNode)
		if !ok {
			return NewTclError(EBADSTATE, "node on operator stack is not an operatorNode!")
		}
		for !top.isSentinel() {
			err := e.reduce()
			if err != nil {
				return err
			}
			top, ok = e.operators.Last().(OperatorNode)
			if !ok {
				return NewTclError(EBADSTATE, "node on operator stack is not an operatorNode!")
			}
		}
	}
	return nil
}

// forcePrecedence will reduce the operator stack if the new node has higher
// precedence than the operators currently on the stack.
func (e *evaluator) forcePrecedence(node OperatorNode) *TclError {
	// If the operator stack is empty, push the new operator.
	// If it has an operator on top, compare the precedence
	// of the two and push the new one if it has lower precedence
	// (or equal precedence: this will force left associativity).
	// Otherwise reduce the two stacks.
	if e.operators.Len() > 0 {
		top, ok := e.operators.Last().(OperatorNode)
		if !ok {
			return NewTclError(EBADSTATE,
				"node on operator stack is not an operatorNode!")
		}
		if !top.isSentinel() && node.getPrecedence() >= top.getPrecedence() {
			err := e.reduce()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Evaluate parses the input string as a Tcl expression, evaluating it and
// returning the result. Supported expressions include variable references,
// nested commands (inside square brackets), string and numeric literals, and
// math and type coercion functions.
func (e *evaluator) Evaluate(expr string) (string, returnCode, *TclError) {
	// reset the evaluator so it is ready for a new expression
	e.state = searchArgument
	e.arguments = nil
	e.operators = nil
	e.funcCount = 0
	e.root = nil
	// lex the input expression into tokens
	c := lexExpr("Evaluate", expr)

	// pull tokens from lexer, building the expression tree
	for tok := range c {
		if tok.typ == tokenError {
			return "", returnError, NewTclError(ELEXER, tok.val)

		} else if tok.typ == tokenEOF {
			err := e.handleEOF()
			if err != nil {
				return "", returnError, err
			}
			break

		} else if tok.typ == tokenVariable || tok.typ == tokenCommand ||
			tok.typ == tokenInteger || tok.typ == tokenFloat ||
			tok.typ == tokenString || tok.typ == tokenBrace ||
			tok.typ == tokenQuote {
			node := newExprNode(e, tok)
			e.arguments.Push(node)
			e.state = searchOperator

		} else if tok.typ == tokenOperator {
			// based on search state, it's either a unary or binary operator
			var node *operatorNode
			if e.state == searchOperator {
				node = newOperatorNode(e, tok, 2)
			} else if e.state == searchArgument {
				node = newOperatorNode(e, tok, 1)
			}
			e.forcePrecedence(node)
			e.operators.Push(node)
			e.state = searchArgument

		} else if tok.typ == tokenFunction {
			// expecting arguments until right parenthesis encountered
			e.funcCount++
			node := newFunctionNode(e, tok)
			// Put it on the argument stack as a sentinel, to mark
			// the beginning of the function arguments.
			e.arguments.Push(node)
			// Put it on the operator stack so when we find the
			// right parenthesis, we can determine that we were
			// making a function call.
			e.operators.Push(node)
			e.state = searchArgument

		} else if tok.typ == tokenParen {
			if tok.val == "(" {
				leftParen := newOperatorNode(e, tok, 1)
				e.operators.Push(leftParen)
			} else {
				// If not open paren, then it is close paren.
				err := e.handleCloseParen()
				if err != nil {
					return "", returnError, err
				}
			}

		} else if tok.typ == tokenComma {
			e.handleComma()
			e.state = searchArgument
		}
	}

	if e.root == nil {
		panic("expression parsing failed!")
	}
	result, code, err := e.root.evaluate()
	if err != nil {
		return "", returnError, err
	}
	return fmt.Sprint(result), code, nil
}

// Interpret passes the expression to the Tcl interpreter associated with
// this evaluator and returns the result.
func (e *evaluator) Interpret(expr string) (string, returnCode, *TclError) {
	return e.interp.Evaluate(expr)
}

// GetVariable retrieves the value for the named variable from the
// interpreter.
func (e *evaluator) GetVariable(name string) (string, *TclError) {
	return e.interp.GetVariable(name)
}
