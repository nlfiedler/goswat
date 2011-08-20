//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// escapes maps an escape code to the matching character literal.
var escapes = map[int]string{'a': "\a", 'b': "\b", 'f': "\f", 'n': "\n", 'r': "\r", 't': "\t", 'v': "\v"}

// coerceNumber attempts to parse the given expression as either
// an integer or floating point number. Failing that, it returns
// the input as given. Leading plus and minus signs are honored.
func coerceNumber(expr string) (interface{}, *TclError) {
	pe := expr
	if len(expr) > 1 && (expr[0] == '-' || expr[0] == '+') {
		pe = expr[1:]
	}
	p := NewParser(pe)
	state, _ := p.parseNumber()
	if state != stateOK {
		// it was not a number
		return expr, nil
	}
	switch p.token {
	case tokenInteger:
		return atoi(expr)
	case tokenFloat:
		return atof(expr)
	default:
		// parser made a mistake
		return "", NewTclError(EINVALNUM, expr)
	}
	panic("unreachable")
}

// atof attempts to coerce the given text into a floating point value,
// returning an error if unsuccessful.
func atof(text string) (interface{}, *TclError) {
	v, err := strconv.Atof64(text)
	if err != nil {
		if err == os.EINVAL {
			// the parser messed up if this happens
			return "", NewTclError(EINVALNUM, text)
		}
		if err == os.ERANGE {
			return "", NewTclError(ENUMRANGE, text)
		}
	}
	return v, nil
}

// atoi attempts to coerce the given text into an integer value,
// returning an error if unsuccessful.
func atoi(text string) (interface{}, *TclError) {
	// let strconv detect the number base for us
	// (either binary, decimal, or hexadecimal)
	v, err := strconv.Btoi64(text, 0)
	if err != nil {
		if err == os.EINVAL {
			// the parser messed up if this happens
			return "", NewTclError(EINVALNUM, text)
		}
		if err == os.ERANGE {
			return "", NewTclError(ENUMRANGE, text)
		}
	}
	return v, nil
}

// evalBoolean attempts to interpret the given string as a boolean
// expression. If expr is a number, 0 means false while all other number
// result in true. If expr is "yes" or "true" then the result is true.
// If expr is "no" or "false" then the result is false. Otherwise an
// error is returned.
func evalBoolean(expr string) (bool, *TclError) {
	n, err := strconv.Atoi(expr)
	if err == nil {
		return n != 0, nil
	}
	s := strings.ToLower(expr)
	if s == "false" || s == "no" {
		return false, nil
	} else if s == "true" || s == "yes" {
		return true, nil
	}
	return false, NewTclError(EBADBOOL, "expected true/false or yes/no")
}

// hexCharToByte converts the given character to a byte value, where the
// character represents a hexadecimal digit (0..9, a..f, A..F). Returns
// -1 if the character is not a valid hex digit.
func hexCharToByte(r int) int {
	if r >= '0' && r <= '9' {
		return r - '0'
	} else if r >= 'a' && r <= 'f' {
		return r - 'a' + 10
	} else if r >= 'A' && r <= 'F' {
		return r - 'A' + 10
	} else {
		return -1
	}
	panic("unreachable")
}

// octCharToByte converts the given character to a byte value, where the
// character represents an octal digit (0..7). Returns -1 if the
// character is not a valid hex digit.
func octCharToByte(r int) int {
	if r >= '0' && r <= '7' {
		return r - '0'
	} else {
		return -1
	}
	panic("unreachable")
}

// evalString performs string substitution on the given expression,
// returning the result. This does not handle interpolation of nested
// commands and variable references.
func evalString(expr string) (string, *TclError) {
	if strings.Index(expr, "\\") == -1 {
		// nothing to do
		return expr, nil
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(expr)))
	escaped := false
	unicode := 0
	hex := 0
	octal := 0
	num := 0
	for _, c := range expr {
		if unicode > 0 {
			v  := hexCharToByte(c)
			if v == -1 {
				return "", NewTclError(EINVALNUM, "invalid hex character in " + expr)
			}
			num = num << 4 + v
			unicode--
			if unicode == 0 {
				buf.WriteRune(num)
				num = 0
			}

		} else if hex > 0 {
			v  := hexCharToByte(c)
			if v == -1 {
				return "", NewTclError(EINVALNUM, "invalid hex character in " + expr)
			}
			num = num << 4 + v
			hex--
			if hex == 0 {
				buf.WriteByte(byte(num))
				num = 0
			}

		} else if octal > 0 {
			v  := octCharToByte(c)
			if v == -1 {
				return "", NewTclError(EINVALNUM, "invalid octal character in " + expr)
			}
			num = num << 3 + v
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

// widenNumber will convert a lower precision numeric value to the
// widest possible precision signed value (i.e. int64, float64). If the
// input is an uint64 then an error is returned since the conversion to
// int64 would lose precision. If the value is not a number at all, an
// error is returned.
// func widenNumber(num interface{}) (interface{}, *TclError) {
// 	if _, ok := num.(uint); ok {
// 		return 0, NewTclError(ENUMRANGE, "cannot widen uint to int64")
// 	}
// 	if i, ok := num.(int); ok {
// 		return int64(i), nil
// 	}
// 	if ui8, ok := num.(uint8); ok {
// 		return int64(ui8), nil
// 	}
// 	if ui16, ok := num.(uint16); ok {
// 		return int64(ui16), nil
// 	}
// 	if ui32, ok := num.(uint32); ok {
// 		return int64(ui32), nil
// 	}
// 	if _, ok := num.(uint64); ok {
// 		return 0, NewTclError(ENUMRANGE, "cannot widen uint64 to int64")
// 	}
// 	if i8, ok := num.(int8); ok {
// 		return int64(i8), nil
// 	}
// 	if i16, ok := num.(int16); ok {
// 		return int64(i16), nil
// 	}
// 	if i32, ok := num.(int32); ok {
// 		return int64(i32), nil
// 	}
// 	if i64, ok := num.(int64); ok {
// 		return i64, nil
// 	}
// 	if f32, ok := num.(float32); ok {
// 		return float64(f32), nil
// 	}
// 	if f64, ok := num.(float64); ok {
// 		return f64, nil
// 	}
// 	return 0, NewTclError(EOPERAND, "not a number")
// }

// performAddition adds to values together. If both values are numbers
// then they will be added (if either is a floating point, the result
// will be a float). Otherwise, the two values are converted to strings
// with the right appended to the left. All numeric values must be the
// widest possible signed type (i.e. int64 or float64).
func performAddition(left, right interface{}) (interface{}, *TclError) {
	lv := left
	rv := right
	if lv == nil || rv == nil {
		return nil, NewTclError(EOPERAND, "cannot operate on nil")
	}
	// if nl, err := widenNumber(lv); err == nil {
	// 	lv = nl
	// }
	// if nr, err := widenNumber(rv); err == nil {
	// 	rv = nr
	// }
	lf, lf_ok := lv.(float64)
	rf, rf_ok := rv.(float64)
	li, li_ok := lv.(int64)
	ri, ri_ok := rv.(int64)
	if lf_ok && rf_ok {
		return lf + rf, nil
	} else if lf_ok && ri_ok {
		return lf + float64(ri), nil
	} else if li_ok && rf_ok {
		return float64(li) + rf, nil
	} else if li_ok && ri_ok {
		return li + ri, nil
	} else {
		// not numbers, just append them as strings
		return fmt.Sprintf("%v%v", lv, rv), nil
	}
	panic("unreachable")
}

// performSubtraction subtracts the right value from the left value. If
// either value is not a number, an error is returned. All numeric
// values must be the widest possible signed type (i.e. int64 or
// float64).
func performSubtraction(left, right interface{}) (interface{}, *TclError) {
	lv := left
	rv := right
	if lv == nil || rv == nil {
		return nil, NewTclError(EOPERAND, "cannot operate on nil")
	}
	// if nl, err := widenNumber(lv); err == nil {
	// 	lv = nl
	// }
	// if nr, err := widenNumber(rv); err == nil {
	// 	rv = nr
	// }
	lf, lf_ok := lv.(float64)
	rf, rf_ok := rv.(float64)
	li, li_ok := lv.(int64)
	ri, ri_ok := rv.(int64)
	if lf_ok && rf_ok {
		return lf - rf, nil
	} else if lf_ok && ri_ok {
		return lf - float64(ri), nil
	} else if li_ok && rf_ok {
		return float64(li) - rf, nil
	} else if li_ok && ri_ok {
		return li - ri, nil
	} else {
		return nil, NewTclError(EOPERAND, "cannot operate on non-numeric values")
	}
	panic("unreachable")
}
