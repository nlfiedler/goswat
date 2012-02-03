//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package liswat

//
// Parser for our Scheme-like language, which turns tokens from the
// lexer into a tree of expressions to be evaluated.
//

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
)

// Symbol represents a variable or procedure name in a Scheme
// expression. It is essentially a string but is treated differently.
type Symbol string

var eofObject Symbol = Symbol("#<eof-object>")

// stringify takes a tree of elements and converts it to a string in
// Scheme format (e.g. true is "#t", lists are "(...)", etc).
func stringify(x interface{}) string {
	buf := new(bytes.Buffer)
	stringifyBuffer(x, buf)
	return buf.String()
}

// stringifyBuffer converts the tree of elements to a string, which is
// written to the given buffer.
func stringifyBuffer(x interface{}, buf *bytes.Buffer) {
	switch i := x.(type) {
	case nil:
		buf.WriteString("()")
	case []interface{}:
		buf.WriteString("(")
		for _, v := range i {
			stringifyBuffer(v, buf)
			buf.WriteString(" ")
		}
		// lop off the trailing space
		if buf.Len() > 2 {
			buf.Truncate(buf.Len() - 1)
		}
		buf.WriteString(")")
	case bool:
		if i {
			buf.WriteString("#t")
		} else {
			buf.WriteString("#f")
		}
	case Symbol:
		buf.WriteString(string(i))
	case string:
		buf.WriteString(fmt.Sprintf("\"%s\"", i))
	default:
		buf.WriteString(fmt.Sprintf("%v", i))
	}
}

// parseExpr parses a Lisp expression and returns the result, which may
// be a string, number, symbol, or a list of expressions.
func parseExpr(expr string) (interface{}, *LispError) {
	c := lex("parseExpr", expr)
	t, ok := <-c
	if !ok {
		return nil, NewLispError(ELEXER, "unexpected end of lexer stream")
	}
	if t.typ == tokenEOF {
		return eofObject, nil
	}
	return parserRead(t, c)
}

// parserRead reads a complete expression from the channel of tokens,
// starting with the initial token value provided.
func parserRead(t token, c chan token) (interface{}, *LispError) {
	switch t.typ {
	case tokenError:
		return nil, NewLispError(ELEXER, t.val)
	case tokenEOF:
		return nil, NewLispError(ELEXER, "unexpected EOF in list")
	case tokenOpenParen:
		list := make([]interface{}, 0)
		for {
			t = <-c
			if t.typ == tokenCloseParen {
				return list, nil
			}
			val, err := parserRead(t, c)
			if err != nil {
				return nil, err
			}
			list = append(list, val)
		}
	case tokenCloseParen:
		return nil, NewLispError(ESYNTAX, "unexpected )")
	case tokenString:
		// TODO: decode string escapes; could use strconv.Unquote(string)
		return t.contents(), nil
	case tokenInteger:
		return atoi(t.val)
	case tokenFloat:
		return atof(t.val)
	case tokenBoolean:
		if t.val == "#t" || t.val == "#T" {
			return true, nil
		} else {
			// lexer already validated that it is #f or #F
			return false, nil
		}
	case tokenCharacter:
		// TODO: what is character used for?
		return nil, NewLispError(ESYNTAX, t.val+" is unsupported")
	case tokenQuote:
		// TODO: handle quotes (' ` , ,@)
		return nil, NewLispError(ESYNTAX, t.val+" is unsupported")
	case tokenIdentifier:
		return Symbol(t.val), nil
	}
	panic("unreachable code")
}

// atof attempts to coerce the given text into a floating point value,
// returning an error if unsuccessful.
func atof(text string) (interface{}, *LispError) {
	v, err := strconv.ParseFloat(text, 64)
	if err != nil {
		if err == os.EINVAL {
			// the lexer messed up if this happens
			return "", NewLispError(EINVALNUM, text)
		}
		if err == os.ERANGE {
			return "", NewLispError(ENUMRANGE, text)
		}
	}
	return v, nil
}

// atoi attempts to coerce the given text into an integer value,
// returning an error if unsuccessful.
func atoi(text string) (interface{}, *LispError) {
	// let strconv detect the number base for us
	// (either binary, decimal, or hexadecimal)
	v, err := strconv.ParseInt(text, 0, 64)
	if err != nil {
		if err == os.EINVAL {
			// the lexer messed up if this happens
			return "", NewLispError(EINVALNUM, text)
		}
		if err == os.ERANGE {
			return "", NewLispError(ENUMRANGE, text)
		}
	}
	return v, nil
}
