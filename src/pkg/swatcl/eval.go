//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"bytes"
	"strconv"
	"strings"
)

// escapes maps an escape code to the matching character literal.
var escapes = map[int]string{'a': "\a", 'b': "\b", 'f': "\f", 'n': "\n", 'r': "\r", 't': "\t", 'v': "\v"}

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
