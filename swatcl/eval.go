//
// Copyright 2011-2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"bytes"
	"strconv"
	"strings"
	"unicode"
)

// escapes maps an escape code to the matching character literal.
var escapes = map[rune]string{'a': "\a", 'b': "\b", 'f': "\f", 'n': "\n", 'r': "\r", 't': "\t", 'v': "\v", '\\': "\\"}

// isAlphaNumeric indicates if the given rune is a letter or number.
func isAlphaNumeric(r rune) bool {
	return unicode.IsDigit(r) || unicode.IsLetter(r)
}

// coerceNumber attempts to parse the given expression as either an integer or
// floating point number. Failing that, it returns the input as given. Leading
// plus and minus signs are honored.
func coerceNumber(expr string) (interface{}, error) {
	// prepare the expression to be lexed
	pe := expr
	if len(expr) > 1 && (expr[0] == '-' || expr[0] == '+') {
		pe = expr[1:]
	}
	c := lexExpr("coerceNumber", pe)
	token := <-c
	if token.typ == tokenFloat {
		return atof(expr)
	} else if token.typ == tokenInteger {
		return atoi(expr)
	} else {
		return expr, nil
	}
	panic("unreachable code")
}

// atof attempts to coerce the given text into a floating point value,
// returning an error if unsuccessful.
func atof(text string) (interface{}, error) {
	v, err := strconv.ParseFloat(text, 64)
	if err != nil {
		if err == strconv.ErrSyntax {
			// the parser messed up if this happens
			return nil, NumberSyntaxInvalid
		}
		if err == strconv.ErrRange {
			return nil, NumberOutOfRange
		}
	}
	return v, nil
}

// atoi attempts to coerce the given text into an integer value, returning an
// error if unsuccessful.
func atoi(text string) (interface{}, error) {
	// let strconv detect the number base for us
	// (either binary, decimal, or hexadecimal)
	v, err := strconv.ParseInt(text, 0, 64)
	if err != nil {
		if err == strconv.ErrSyntax {
			// the parser messed up if this happens
			return nil, NumberSyntaxInvalid
		}
		if err == strconv.ErrRange {
			return nil, NumberOutOfRange
		}
	}
	return v, nil
}

// evalBoolean attempts to interpret the given string as a boolean expression.
// If expr is a number, 0 means false while all other numbers result in true.
// If expr is "on", "yes", or "true" then the result is true. If expr is
// "off", "no", or "false" then the result is false. Otherwise an error is
// returned.
func evalBoolean(expr string) (bool, error) {
	n, err := strconv.Atoi(expr)
	if err == nil {
		return n != 0, nil
	}
	s := strings.ToLower(expr)
	if s == "false" || s == "no" || s == "off" {
		return false, nil
	} else if s == "true" || s == "yes" || s == "on" {
		return true, nil
	}
	return false, BooleanInvalidSyntax
}

// hexCharToByte converts the given character to a byte value, where the
// character represents a hexadecimal digit (0..9, a..f, A..F). Returns -1 if
// the character is not a valid hex digit.
func hexCharToByte(r rune) rune {
	if r >= '0' && r <= '9' {
		return r - '0'
	} else if r >= 'a' && r <= 'f' {
		return r - 'a' + 10
	} else if r >= 'A' && r <= 'F' {
		return r - 'A' + 10
	} else {
		return -1
	}
	panic("unreachable code")
}

// octCharToByte converts the given character to a byte value, where the
// character represents an octal digit (0..7). Returns -1 if the character is
// not a valid hex digit.
func octCharToByte(r rune) rune {
	if r >= '0' && r <= '7' {
		return r - '0'
	} else {
		return -1
	}
	panic("unreachable code")
}

// evalString performs string substitution on the given expression, returning
// the result. This does not handle interpolation of nested commands and
// variable references.
func evalString(expr string) (string, error) {
	if strings.Index(expr, "\\") == -1 {
		// nothing to do
		return expr, nil
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(expr)))
	escaped := false
	unicode := 0
	hex := 0
	octal := 0
	var num int32 = 0
	for _, c := range expr {
		if unicode > 0 {
			v := hexCharToByte(c)
			if v == -1 {
				return "", NumberSyntaxInvalid
			}
			num = num<<4 + v
			unicode--
			if unicode == 0 {
				buf.WriteRune(num)
				num = 0
			}

		} else if hex > 0 {
			v := hexCharToByte(c)
			if v == -1 {
				return "", NumberSyntaxInvalid
			}
			num = num<<4 + v
			hex--
			if hex == 0 {
				buf.WriteByte(byte(num))
				num = 0
			}

		} else if octal > 0 {
			v := octCharToByte(c)
			if v == -1 {
				return "", NumberSyntaxInvalid
			}
			num = num<<3 + v
			octal--
			if octal == 0 {
				buf.WriteByte(byte(num))
				num = 0
			}

		} else if escaped {
			switch c {
			case 'u':
				unicode = 4
			case 'x':
				hex = 2
			case '0':
				octal = 2
			default:
				if v, ok := escapes[c]; ok {
					buf.WriteByte(v[0])
				}
			}
			escaped = false

		} else if c == '\\' {
			escaped = true
		} else {
			buf.WriteRune(c)
		}
	}
	return buf.String(), nil
}
