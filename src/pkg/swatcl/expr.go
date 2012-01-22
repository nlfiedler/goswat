//
// Copyright 2011 Nathan Fiedler. All rights reserved.
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
	"container/vector"
	"fmt"
)

//
// TODO: write a lexer for Tcl expressions, similar to lexer.go, using
// the parseExprToken as an example of what to do, and replace the
// searchState below with lexer state functions; could reuse the lexer
// in lexer.go if we can provide our own initial state function in the
// call to the lex() function
//

// TODO: support variable references in expressions

// TODO: support command invocations in expressions

// TODO: support following math functions
// abs         acos        asin        atan
// atan2       bool        ceil        cos
// cosh        double      entier      exp
// floor       fmod        hypot       int
// isqrt       log         log10       max
// min         pow         rand        round
// sin         sinh        sqrt        srand
// tan         tanh        wide
// (where double, int, wide, entier are type conversions)

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

// Evaluator knows how to evaluate a Tcl expression.
type Evaluator interface {
	evaluate() (interface{}, *TclError)
}

// exprNode contains the attributes of an expression node, whether it
// is a numeric or string literal, or a command or variable reference.
type exprNode struct {
	token parserToken
	text  string
	eval  *evaluator
}

// newExprNode creates a new expression node based on the token and text.
func newExprNode(eval *evaluator, token parserToken, text string) *exprNode {
	return &exprNode{token, text, eval}
}

// evaluate evaluates the expression node appropriately based on its type.
func (n *exprNode) evaluate() (interface{}, *TclError) {
	switch n.token {
	case tokenVariable:
		str, err := n.eval.interp.GetVariable(n.text)
		if err != nil {
			return "", err
		}
		val, err := coerceNumber(str)
		if err != nil {
			return "", err
		}
		return val, nil
	case tokenCommand:
		return "[cmd]", nil // TODO: implement command invocation
	case tokenInteger:
		return atoi(n.text)
	case tokenFloat:
		return atof(n.text)
	case tokenString:
		// perform basic string substitution (slash escapes)
		return evalString(n.text)
	case tokenBrace:
		// return the string as-is
		return n.text, nil
	}
	return "", nil
}

type functionNode struct {
	exprNode
	arguments vector.Vector // function arguments
}

func (f *functionNode) evaluate() {
	// TODO: evaluate the arguments and invoke the function
}

type evaluator struct {
	state     searchState
	root      Evaluator     // root of the expression tree
	interp    *Interpreter  // contains program state
	arguments vector.Vector // argument stack
	operators vector.Vector // operator stack
	funcCount int
	//prevToken parserToken
}

func newEvaluator(interp *Interpreter) *evaluator {
	e := &evaluator{
		state: searchArgument,
		interp: interp,
	}
	return e
}

// Reduce the operator stack by one. If the operator stack top is a left
// parenthesis, no change is made.
func (e *evaluator) reduce() *TclError {
	// If there is a binary operator on top of the operator stack,
	// there should be two trees on top of the argument stack, both
	// representing expressions. Pop the operator and two trees off
	// of the argument stack, combining them into a single tree
	// node, which is then pushed back on the argument stack. Note
	// that the trees on the argument stack represent the right and
	// left arguments, respectively.
	top, ok := e.operators.Pop().(*operatorNode)
	if !ok {
		return NewTclError(EBADSTATE, "node on operator stack is not an operator!")
	}
	if top.sentinel {
		// Cleverly do nothing and let the caller handle it.
	} else if top.arity == 2 {
		if e.arguments.Len() < 2 {
			return NewTclError(EOPERAND, "operator requires two arguments")
		}
		arg2, ok := e.arguments.Pop().(*exprNode)
		if !ok {
			return NewTclError(EBADSTATE, "second argument is not an exprNode!")
		}
		arg1, ok := e.arguments.Pop().(*exprNode)
		if !ok {
			return NewTclError(EBADSTATE, "first argument is not an exprNode!")
		}
		top.left = arg1
		top.right = arg2
		e.arguments.Push(top)
	} else if top.arity == 1 {
		if e.arguments.Len() < 1 {
			return NewTclError(EOPERAND, "operator requires one argument")
		}
		arg, ok := e.arguments.Pop().(*exprNode)
		if !ok {
			return NewTclError(EBADSTATE, "single argument is not an exprNode!")
		}
		top.right = arg
		e.arguments.Push(top)
	} else {
		return NewTclError(EOPERATOR, "unknown operator "+top.text)
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
		top, ok := e.operators.Last().(*operatorNode)
		if !ok {
			return NewTclError(EBADSTATE, "node on operator stack is not an operatorNode!")
		}
		// TODO: handle these error cases
		// if (top instanceof LeftParen) {
		//     setError(Errors.UNMATCHED_LPAREN, top.getToken());
		//     return;
		// } else if (top instanceof LeftBracket) {
		//     setError(Errors.UNMATCHED_LBRACKET, top.getToken());
		//     return;
		// } else
		if top.sentinel {
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
		topArg, ok := e.arguments.Pop().(Evaluator)
		if !ok {
			return NewTclError(EBADSTATE, "node on argument stack is not an Evaluator!")
		}
		if e.arguments.Len() == 0 && e.operators.Len() == 0 {
			e.root = topArg
		} else {
			return NewTclError(EBADSTATE, "argument stack is not empty")
		}
	}
	return nil
}

// @Override
// public void caseTLParenthese(TLParenthese node) {
//     // Append state is definitely wrong, but argument and operator
//     // states are perfectly exceptable.
//     if (searchState == State.APPEND) {
//         setError(Errors.DOT_REQUIRES_ID, node);
//         return;
//     }
//     LeftParen lp = new LeftParen(node);
//     // If the last token was an identifier, then this is a method call.
//     if (previousToken instanceof TIdentifier) {
//         // The argument stack is assumed to be empty.
//         Node n = argumentStack.pop();
//         MethodNode method = null;
//         if (n instanceof IdentifierNode) {
//             IdentifierNode inode = (IdentifierNode) n;
//             method = new MethodNode(inode.getToken(), inode.getIdentifier());
//         } else {
//             JoinOperatorNode onode = (JoinOperatorNode) n;
//             Node object = onode.getChild(0);
//             String name = onode.getChild(1).getToken().getText();
//             method = new MethodNode(onode.getToken(), object, name);
//         }
//         // Put it on the argument stack as a sentinel, to mark
//         // the beginning of the method arguments.
//         argumentStack.push(method);
//         // Put it on the operator stack so when we find the
//         // right parenthesis, we can determine that we were
//         // making a method call.
//         operatorStack.push(method);
//         searchState = State.ARGUMENT;
//         methodCount++;
//     }
//     // Else, it is the start of a type-cast or a subgroup.
//     operatorStack.push(lp);
// }

// @Override
// public void caseTRParenthese(TRParenthese node) {
//     if (operatorStack.empty()) {
//         setError(Errors.UNMATCHED_RPAREN, node);
//         return;
//     }
//     // If there is a left parenthesis on the operator stack, we can
//     // "cancel" the pair. If the operator stack contains some other
//     // operator on top, reduce the stacks. This also covers the case
//     // where the parentheses were used for grouping only.
//     OperatorNode top = (OperatorNode) operatorStack.peek();
//     while (!(top instanceof LeftParen)) {
//         reduce();
//         if (operatorStack.empty() || top.isSentinel()) {
//             setError(Errors.UNMATCHED_RPAREN, node);
//             return;
//         }
//         top = (OperatorNode) operatorStack.peek();
//     }
//     operatorStack.pop();

//     // Now check for the method invocation case.
//     if (!operatorStack.empty()
//         && operatorStack.peek() instanceof MethodNode) {
//         // It was a method invocation.
//         MethodNode method = (MethodNode) operatorStack.pop();
//         // Pop off the arguments and add them in reverse order.
//         Node n = argumentStack.pop();
//         Stack<Node> args = new Stack<Node>();
//         while (n != method) {
//             args.push(n);
//             n = argumentStack.pop();
//         }
//         while (!args.empty()) {
//             Node arg = args.pop();
//             method.addChild(arg);
//         }
//         // Put the method invocation back on the argument stack
//         // because it is treated as a value, not an operator.
//         argumentStack.push(method);
//         methodCount--;

//     } else {
//         // Maybe it is a type-cast operation; otherwise it was a
//         // grouping operator and that has been taken care of.
//         try {
//             Node n = argumentStack.peek();
//             if (n instanceof TypeNode) {
//                 argumentStack.pop();
//                 TypeNode tn = (TypeNode) n;
//                 TypeCastOperatorNode tcon = new TypeCastOperatorNode(
//                     tn.getToken(), tn.getTypeName());
//                 handleOperator(tcon);
//             } else if (n instanceof IdentifierNode) {
//                 argumentStack.pop();
//                 IdentifierNode in = (IdentifierNode) n;
//                 TypeCastOperatorNode tcon = new TypeCastOperatorNode(
//                     in.getToken(), in.getIdentifier());
//                 handleOperator(tcon);
//             } else if (n instanceof JoinOperatorNode) {
//                 argumentStack.pop();
//                 JoinOperatorNode jon = (JoinOperatorNode) n;
//                 TypeCastOperatorNode tcon = new TypeCastOperatorNode(
//                     jon.getToken(), jon.mergeChildren());
//                 handleOperator(tcon);
//             }
//         } catch (EvaluationException ee) {
//             setError(ee.getMessage(), top.getToken());
//         } catch (EmptyStackException ese) {
//             setError(Errors.MISSING_ARGS, top.getToken());
//         }
//     }
// }

// @Override
// public void caseTLBracket(TLBracket node) {
//     // Make sure there is something reasonable on the stack, since
//     // a left bracket without a preceding type or identifier is
//     // incorrect syntax.
//     if (argumentStack.isEmpty()) {
//         setError(Errors.UNEXPECTED_TOKEN, node);
//     } else {
//         Node n = argumentStack.peek();
//         if (!(n instanceof JoinOperatorNode)
//             && !(n instanceof IdentifierNode)
//             && !(n instanceof TypeNode)) {
//             setError(Errors.UNEXPECTED_TOKEN, node);
//         }
//     }
//     // Push the left bracket on the operator stack. We can't tell yet
//     // if this is going to be a typecast or an array access.
//     Node lb = new LeftBracket(node);
//     operatorStack.push(lb);
//     // Put it on the argument stack as a sentinel for the array index,
//     // if that is indeed what this turns out to be.
//     argumentStack.push(lb);
//     // We may have been looking for an argument or an operator,
//     // and we can't be sure what we will find next.
// }

// @Override
// public void caseTRBracket(TRBracket node) {
//     // If there is a left bracket on the operator stack, we can
//     // "cancel" the pair. If the operator stack contains some other
//     // operator on top, reduce the stacks.
//     if (operatorStack.empty()) {
//         setError(Errors.UNMATCHED_RBRACKET, node);
//         return;
//     }
//     OperatorNode top = (OperatorNode) operatorStack.peek();
//     while (!(top instanceof LeftBracket)) {
//         reduce();
//         if (operatorStack.empty() || top.isSentinel()) {
//             setError(Errors.UNMATCHED_RBRACKET, node);
//             return;
//         }
//         top = (OperatorNode) operatorStack.peek();
//     }
//     operatorStack.pop();

//     // Was there anything between the brackets?
//     // (we know there will be a left bracket and something else on the
//     // argument stack, given that we are here).
//     Node index = argumentStack.pop();
//     Node name = argumentStack.pop();
//     if (index instanceof LeftBracket) {
//         // It was probably part of a typecast, but we can't be sure.
//         // In any case, make a TypeNode out of this and put it on
//         // the argument stack.
//         Token token = name.getToken();
//         argumentStack.push(new TypeNode(token, token.getText() + "[]"));
//     } else {
//         // It was an array reference.
//         // Make sure that only one array index was provided.
//         if (name instanceof LeftBracket) {
//             ArrayNode arrayref = new ArrayNode(name.getToken());
//             // Retrieve the thing that was there before the left bracket
//             // (an identifier or type).
//             name = argumentStack.pop();
//             arrayref.addChild(name);
//             arrayref.addChild(index);
//             // Put the array reference on the argument stack; it's a value.
//             argumentStack.push(arrayref);
//         } else {
//             setError(Errors.ARRAY_MULTI_INDEX, name.getToken());
//         }
//     }
// }

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
		top, ok := e.operators.Last().(*operatorNode)
		if !ok {
			return NewTclError(EBADSTATE, "node on operator stack is not an operatorNode!")
		}
		for !top.sentinel {
			err := e.reduce()
			if err != nil {
				return err
			}
			top, ok = e.operators.Last().(*operatorNode)
			if !ok {
				return NewTclError(EBADSTATE, "node on operator stack is not an operatorNode!")
			}
		}
	}
	return nil
}

// TODO: see the other relevant case methods in TreeBuilder.java

// EvaluateExpression parses the input string as a Tcl expression,
// evaluating it and returning the result. Supported expressions include
// variable references, nested commands (inside square brackets), string
// and numeric literals, and math and type coercion functions.
func EvaluateExpression(interp *Interpreter, expr string) (string, *TclError) {
	p := NewParser(expr)
	e := newEvaluator(interp)

	// TODO: get the evaluator working with operator precedence (e.g. * before +)
	// TODO: get the evaluator working for grouped expressions (e.g. (1 + 2) * 3)
	// TODO: get the evaluator working for variable expressions
	// TODO: get the evaluator working for nested commands
	// TODO: get the evaluator working for function invocation

	for {
		// pull tokens from parser, building expression tree
		p.parseExprToken()
		if p.token == tokenEOF {
			err := e.handleEOF()
			if err != nil {
				return "", err
			}
			break
		}
		t := p.GetTokenText()
		if p.token == tokenVariable || p.token == tokenCommand ||
			p.token == tokenInteger || p.token == tokenFloat ||
			p.token == tokenString || p.token == tokenBrace {
			node := newExprNode(e, p.token, t)
			e.arguments.Push(node)
			e.state = searchOperator

		} else if p.token == tokenOperator {
			// based on search state, it's either a unary or binary operator
			var node *operatorNode
			if e.state == searchOperator {
				node = newOperatorNode(e, p.token, t, 2)
			} else if e.state == searchArgument {
				node = newOperatorNode(e, p.token, t, 1)
			}

			// If the operator stack is empty, push the new operator.
			// If it has an operator on top, compare the precedence
			// of the two and push the new one if it has lower precedence
			// (or equal precedence: this will force left associativity).
			// Otherwise reduce the two stacks.
			if e.operators.Len() > 0 {
				top, ok := e.operators.Last().(*operatorNode)
				if !ok {
					return "", NewTclError(EBADSTATE,
						"node on operator stack is not an operatorNode!")
				}
				if !top.sentinel && node.precedence >= top.precedence {
					err := e.reduce()
					if err != nil {
						return "", err
					}
				}
			}
			e.operators.Push(node)
			e.state = searchArgument

		} else if p.token == tokenFunction {
			// TODO: expecting arguments until right parenthesis encountered
			e.funcCount++
			// TODO: should be constructing functionNode here
			node := newExprNode(e, p.token, t)
			// Put it on the argument stack as a sentinel, to mark
			// the beginning of the method arguments.
			e.arguments.Push(node)
			// Put it on the operator stack so when we find the
			// right parenthesis, we can determine that we were
			// making a function call.
			e.operators.Push(node)
			e.state = searchArgument

		} else if p.token == tokenParen {
			// TODO: grouping operator, call caseTLParenthese

		} else if p.token == tokenSeparator {
			// Not finished parsing, continue
			continue
		} else if p.token == tokenComma {
			e.handleComma()
		}
	}

	if e.root == nil {
		panic("expression parsing failed!")
	}
	result, err := e.root.evaluate()
	if err != nil {
		return "", err
	}
	return fmt.Sprint(result), nil
}
