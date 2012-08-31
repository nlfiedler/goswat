//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"bytes"
	"fmt"
)

// commandExpr implements the Tcl 'expr' command.
func commandExpr(i Interpreter, argv []string, data []string) (string, returnCode, *TclError) {
	buf := new(bytes.Buffer)
	for ii := 1; ii < len(argv); ii++ {
		buf.WriteString(argv[ii])
		buf.WriteRune(' ')
	}
	input := buf.String()
	eval := newEvaluator(i)
	return eval.Evaluate(input)
}

// commandIf implements the Tcl 'if/then/elseif/else' command.
func commandIf(i Interpreter, argv []string, data []string) (string, returnCode, *TclError) {
	// if expr1 ?then? body1 elseif expr2 ?then? body2 elseif ... ?else? ?bodyN?
	if len(argv) != 3 && len(argv) != 5 {
		return "", returnError, arityError(argv[0])
	}
	// TODO: allow optional 'then' keyword
	eval := newEvaluator(i)
	result, code, err := eval.Evaluate(argv[1])
	if err != nil {
		return "", code, err
	}
	// TODO: support additional elseif/then clauses
	b, err := evalBoolean(result)
	if err != nil {
		return "", returnError, err
	}
	if b {
		result, code, err = i.Evaluate(argv[2])
	} else if len(argv) == 5 {
		if argv[3] != "else" {
			return "", returnError, NewTclError(ECOMMAND, "missing 'else' keyword prior to last body")
		}
		// TODO: need to check that second last argument is 'else'
		result, code, err = i.Evaluate(argv[4])
	}
	if err != nil {
		return "", returnError, err
	}
	return result, code, nil
}

// commandPuts implements the Tcl 'puts' command (print a string to the console).
func commandPuts(i Interpreter, argv []string, data []string) (string, returnCode, *TclError) {
	if len(argv) < 2 {
		return "", returnError, arityError(argv[0])
	}
	format := "%s\n"
	argi := 1
	if argv[1] == "-nonewline" {
		format = "%s"
		argi = 2
	}
	fmt.Fprintf(i, format, argv[argi])
	return argv[argi], returnOk, nil
}

// commandSet implements the Tcl 'set' command (set a variable value).
func commandSet(i Interpreter, argv []string, data []string) (string, returnCode, *TclError) {
	if len(argv) < 2 {
		return "", returnError, arityError(argv[0])
	}
	if len(argv) == 3 {
		err := i.SetVariable(argv[1], argv[2])
		if err != nil {
			return "", returnError, err
		}
		return argv[2], returnOk, nil
	} else {
		val, err := i.GetVariable(argv[1])
		if err != nil {
			return "", returnError, err
		}
		return val, returnOk, nil
	}
	panic("unreachable code")
}

//
// TODO: translate to idiomatic Go code
//

// int picolCommandWhile(struct picolInterp *i, int argc, char **argv, void *pd) {
//     if (argc != 3) return picolArityErr(i,argv[0]);
//     while(1) {
//         int retcode = picolEval(i,argv[1]);
//         if (retcode != PICOL_OK) return retcode;
//         if (atoi(i->result)) {
//             if ((retcode = picolEval(i,argv[2])) == PICOL_CONTINUE) continue;
//             else if (retcode == PICOL_OK) continue;
//             else if (retcode == PICOL_BREAK) return PICOL_OK;
//             else return retcode;
//         } else {
//             return PICOL_OK;
//         }
//     }
// }

// int picolCommandRetCodes(struct picolInterp *i, int argc, char **argv, void *pd) {
//     if (argc != 1) return picolArityErr(i,argv[0]);
//     if (strcmp(argv[0],"break") == 0) return PICOL_BREAK;
//     else if (strcmp(argv[0],"continue") == 0) return PICOL_CONTINUE;
//     return PICOL_OK;
// }

// int picolCommandCallProc(struct picolInterp *i, int argc, char **argv, void *pd) {
//     char **x=pd, *alist=x[0], *body=x[1], *p=strdup(alist), *tofree;
//     struct picolCallFrame *cf = malloc(sizeof(*cf));
//     int arity = 0, done = 0, errcode = PICOL_OK;
//     char errbuf[1024];
//     cf->vars = NULL;
//     cf->parent = i->callframe;
//     i->callframe = cf;
//     tofree = p;
//     while(1) {
//         char *start = p;
//         while(*p != ' ' && *p != '\0') p++;
//         if (*p != '\0' && p == start) {
//             p++; continue;
//         }
//         if (p == start) break;
//         if (*p == '\0') done=1; else *p = '\0';
//         if (++arity > argc-1) goto arityerr;
//         picolSetVar(i,start,argv[arity]);
//         p++;
//         if (done) break;
//     }
//     free(tofree);
//     if (arity != argc-1) goto arityerr;
//     errcode = picolEval(i,body);
//     if (errcode == PICOL_RETURN) errcode = PICOL_OK;
//     picolDropCallFrame(i); /* remove the called proc callframe */
//     return errcode;
// arityerr:
//     snprintf(errbuf,1024,"Proc '%s' called with wrong arg num",argv[0]);
//     picolSetResult(i,errbuf);
//     picolDropCallFrame(i); /* remove the called proc callframe */
//     return PICOL_ERR;
// }

// int picolCommandProc(struct picolInterp *i, int argc, char **argv, void *pd) {
//     char **procdata = malloc(sizeof(char*)*2);
//     if (argc != 4) return picolArityErr(i,argv[0]);
//     procdata[0] = strdup(argv[2]); /* arguments list */
//     procdata[1] = strdup(argv[3]); /* procedure body */
//     return picolRegisterCommand(i,argv[1],picolCommandCallProc,procdata);
// }

// int picolCommandReturn(struct picolInterp *i, int argc, char **argv, void *pd) {
//     if (argc != 1 && argc != 2) return picolArityErr(i,argv[0]);
//     picolSetResult(i, (argc == 2) ? argv[1] : "");
//     return PICOL_RETURN;
// }
