//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"bytes"
	"fmt"
	"strings"
)

// arityError is a convenience method for commands to report an error with the
// number of arguments given to the command.
func arityError(name string) TclResult {
	return newTclResultErrorf(EARGUMENT, "Wrong number of arguments for '%s'", name)
}

// commandBreak implements the Tcl 'break' command.
func commandBreak(i Interpreter, argv []string, data []string) TclResult {
	if len(argv) != 1 {
		return arityError(argv[0])
	}
	return newTclResult(EOK, "", returnBreak, "")
}

// commandContinue implements the Tcl 'continue' command.
func commandContinue(i Interpreter, argv []string, data []string) TclResult {
	if len(argv) != 1 {
		return arityError(argv[0])
	}
	return newTclResult(EOK, "", returnContinue, "")
}

// commandExpr implements the Tcl 'expr' command.
func commandExpr(i Interpreter, argv []string, data []string) TclResult {
	buf := new(bytes.Buffer)
	for ii := 1; ii < len(argv); ii++ {
		buf.WriteString(argv[ii])
		buf.WriteRune(' ')
	}
	input := buf.String()
	eval := newEvaluator(i)
	return eval.Evaluate(input)
}

// commandIf implements the Tcl 'if/then/elseif/else' commands.
func commandIf(i Interpreter, argv []string, data []string) TclResult {
	// if expr1 ?then? body1 elseif expr2 ?then? body2 elseif ... ?else? ?bodyN?
	if len(argv) != 3 && len(argv) != 5 {
		return arityError(argv[0])
	}
	// TODO: allow optional 'then' keyword
	eval := newEvaluator(i)
	// TODO: can this support the math func/op commands directly?
	result := eval.Evaluate(argv[1])
	if !result.Ok() {
		return result
	}
	// TODO: support additional elseif/then clauses
	b, err := evalBoolean(result.Result())
	if err != nil {
		return newTclResultError(ESYNTAX, err.Error())
	}
	if b {
		result = i.Evaluate(argv[2])
	} else if len(argv) == 5 {
		if argv[3] != "else" {
			return newTclResultError(ECOMMAND, "missing 'else' keyword prior to last body")
		}
		// TODO: need to check that second last argument is 'else'
		result = i.Evaluate(argv[4])
	}
	return result
}

// commandPuts implements the Tcl 'puts' command (print a string to the console).
func commandPuts(i Interpreter, argv []string, data []string) TclResult {
	if len(argv) > 1 {
		format := "%s\n"
		argi := 1
		if argv[1] == "-nonewline" {
			format = "%s"
			argi = 2
		}
		fmt.Fprintf(i, format, argv[argi])
		return newTclResultOk(argv[argi])
	}
	return newTclResultOk("")
}

// returnCodes maps the string representation of return codes (e.g. "break")
// to their numeric values.
var returnCodes = make(map[string]returnCode)

// populateReturnCodes adds mappings to the returnCodes table.
func populateReturnCodes() {
	returnCodes["ok"] = returnOk
	returnCodes["error"] = returnError
	returnCodes["return"] = returnReturn
	returnCodes["break"] = returnBreak
	returnCodes["continue"] = returnContinue
}

// commandReturn implements the Tcl command 'return'.
func commandReturn(i Interpreter, argv []string, data []string) TclResult {
	if len(returnCodes) == 0 {
		populateReturnCodes()
	}
	// TODO: return cmd: support -errorcode, -errorinfo, -level, -options
	var rcode returnCode = returnReturn
	var ecode ErrorCode = EOK
	result := ""
	if len(argv) > 1 {
		for ii := 1; ii < len(argv); ii++ {
			switch argv[ii] {
			case "-code":
				ii++
				var ok bool
				rcode, ok = returnCodes[argv[ii]]
				if !ok {
					return newTclResultErrorf(EARGUMENT,
						"return: unknown return code: %s", argv[ii])
				}
				if rcode == returnError {
					ecode = ECOMMAND
				}
			default:
				result = argv[ii]
			}
		}
	}
	return newTclResult(ecode, "", rcode, result)
}

// commandSet implements the Tcl 'set' command (set a variable value).
func commandSet(i Interpreter, argv []string, data []string) TclResult {
	if len(argv) < 2 {
		return arityError(argv[0])
	}
	// TODO: support array element reference as with array(index)
	if len(argv) == 3 {
		err := i.SetVariable(argv[1], argv[2])
		if err != nil {
			return newTclResultError(EBADSTATE, err.Error())
		}
		return newTclResultOk(argv[2])
	} else {
		val, err := i.GetVariable(argv[1])
		if err != nil {
			return newTclResultError(EBADSTATE, err.Error())
		}
		return newTclResultOk(val)
	}
	panic("unreachable code")
}

// commandWhile implements the Tcl command 'while'.
func commandWhile(i Interpreter, argv []string, data []string) TclResult {
	if len(argv) != 3 {
		return arityError(argv[0])
	}
	eval := newEvaluator(i)
	for {
		// TODO: can this support the math func/op commands directly?
		result := eval.Evaluate(argv[1])
		if !result.Ok() {
			return result
		}
		test, err := evalBoolean(result.Result())
		if err != nil {
			return newTclResultError(ESYNTAX, err.Error())
		}
		if test {
			result = i.Evaluate(argv[2])
			if !result.Ok() {
				return result
			}
			if result.ReturnCode() == returnBreak {
				break
			}
		} else {
			break
		}
	}
	return newTclResultOk("")
}

// invokeProcedure calls the previously registered user-defined procedure and
// returns the results. It adds a new stack frame to the interpter, sets
// variables using the names defined as the argument list, evaluates the
// procedure body, and then removes the stack frame.
func invokeProcedure(i Interpreter, argv []string, data []string) TclResult {
	if len(data) != 2 {
		return newTclResultErrorf(EARGUMENT,
			"registered proc '%s' missing private data", argv[0])
	}
	args := strings.Split(data[0], " ")
	if len(args)+1 != len(argv) {
		return arityError(argv[0])
	}
	i.addFrame()
	ii := 1
	for _, name := range args {
		i.SetVariable(strings.TrimSpace(name), argv[ii])
		ii++
	}
	result := i.Evaluate(data[1])
	i.popFrame()
	return result
}

// commandProc implements the Tcl 'proc' command.
func commandProc(i Interpreter, argv []string, data []string) TclResult {
	if len(argv) != 4 {
		return arityError(argv[0])
	}
	privdata := make([]string, 0, 2)
	privdata = append(privdata, argv[2], argv[3])
	i.RegisterCommand(argv[1], invokeProcedure, privdata)
	return newTclResultOk("")
}
