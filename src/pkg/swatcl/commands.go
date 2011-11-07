//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"fmt"
)

// commandIf implements the Tcl 'if/then/elseif/else' command.
func commandIf(i *Interpreter, argv []string, data []string) (parserState, string) {
	// if expr1 ?then? body1 elseif expr2 ?then? body2 elseif ... ?else? ?bodyN?
	if len(argv) != 3 && len(argv) != 5 {
		return i.arityError(argv[0]), ""
	}
	// TODO: allow optional 'then' keyword
	state, err := i.Evaluate(argv[1])
	if state != stateOK {
		return state, err.String()
	}
	// TODO: support additional elseif/then clauses
	b, err := evalBoolean(i.result)
	if err != nil {
		return stateError, err.String()
	}
	if b {
		state, err = i.Evaluate(argv[2])
	} else if len(argv) == 5 {
		if argv[3] != "else" {
			return stateError, "missing 'else' keyword prior to last body"
		}
		// TODO: need to check that second last argument is 'else'
		state, err = i.Evaluate(argv[4])
	}
	if err != nil {
		return stateError, err.String()
	}
	return state, i.result
}

// commandPuts implements the Tcl 'puts' command (print a string to the console).
func commandPuts(i *Interpreter, argv []string, data []string) (parserState, string) {
	if len(argv) != 2 {
		return i.arityError(argv[0]), ""
	}
	fmt.Printf("%s\n", argv[1])
	return stateOK, argv[1]
}

// commandSet implements the Tcl 'set' command (set a variable value).
func commandSet(i *Interpreter, argv []string, data []string) (parserState, string) {
	if len(argv) < 2 {
		return i.arityError(argv[0]), ""
	}
	if len(argv) == 3 {
		err := i.SetVariable(argv[1], argv[2])
		if err != nil {
			return stateError, err.Error()
		}
		return stateOK, argv[2]
	} else {
		val, err := i.GetVariable(argv[1])
		if err != nil {
			return stateError, err.Error()
		}
		return stateOK, val
	}
	panic("unreachable code reached")
}

//
// TODO: translate to idiomatic Go code
//

// int picolCommandMath(struct picolInterp *i, int argc, char **argv, void *pd) {
//     char buf[64]; int a, b, c;
//     if (argc != 3) return picolArityErr(i,argv[0]);
//     a = atoi(argv[1]); b = atoi(argv[2]);
//     if (argv[0][0] == '+') c = a+b;
//     else if (argv[0][0] == '-') c = a-b;
//     else if (argv[0][0] == '*') c = a*b;
//     else if (argv[0][0] == '/') c = a/b;
//     else if (argv[0][0] == '>' && argv[0][1] == '\0') c = a > b;
//     else if (argv[0][0] == '>' && argv[0][1] == '=') c = a >= b;
//     else if (argv[0][0] == '<' && argv[0][1] == '\0') c = a < b;
//     else if (argv[0][0] == '<' && argv[0][1] == '=') c = a <= b;
//     else if (argv[0][0] == '=' && argv[0][1] == '=') c = a == b;
//     else if (argv[0][0] == '!' && argv[0][1] == '=') c = a != b;
//     else c = 0; /* I hate warnings */
//     snprintf(buf,64,"%d",c);
//     picolSetResult(i,buf);
//     return PICOL_OK;
// }

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

// void picolDropCallFrame(struct picolInterp *i) {
//     struct picolCallFrame *cf = i->callframe;
//     struct picolVar *v = cf->vars, *t;
//     while(v) {
//         t = v->next;
//         free(v->name);
//         free(v->val);
//         free(v);
//         v = t;
//     }
//     i->callframe = cf->parent;
//     free(cf);
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
