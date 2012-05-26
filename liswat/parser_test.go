//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package liswat

import (
	"fmt"
	"strings"
	"testing"
)

type expectedExpandError struct {
	err string // expected error message substring
	msg string // explanation of error condition
}

func verifyExpandMap(mapping map[string]string, t *testing.T) {
	for input, expected := range mapping {
		x, err := parseExpr(input)
		if err != nil {
			t.Fatalf("failed to parse %q: %s", input, err)
		}
		_, ok := x.(*Pair)
		if !ok {
			t.Fatalf("parser returned non-pair for %q", input)
		}
		x, err = expand(x, true)
		if err != nil {
			t.Fatalf("failed to expand %q: %s", input, err)
		}
		_, ok = x.(*Pair)
		if !ok {
			t.Fatalf("expand returned non-pair for %q: %T", input, x)
		}
		s := stringify(x)
		if s != expected {
			t.Errorf(`expected <<%s>> but got <<%s>>`, expected, s)
		}
	}
}

func verifyExpandError(t *testing.T, expected map[string]expectedExpandError) {
	for input, error := range expected {
		x, err := parseExpr(input)
		if err != nil {
			t.Fatalf("failed to parse %q: %s", input, err)
		}
		x, err = expand(x, true)
		if err == nil {
			t.Fatalf("expand() should have failed with %q", input)
		}
		if !strings.Contains(err.String(), error.err) {
			t.Errorf("expected [%s] but got [%s] for input %q", error.err, err, input)
		}
	}
}

func verifyParse(input, expected string, t *testing.T) {
	result, err := parseExpr(input)
	if err != nil {
		msg := fmt.Sprintf("failed to parse expression '%s', %s", input, err)
		t.Errorf(msg)
	} else {
		actual := stringify(result)
		if actual != expected {
			t.Errorf("expected <<%s>>, but got <<%s>>", expected, actual)
		}
	}
}

func verifyParseMap(mapping map[string]string, t *testing.T) {
	for input, expected := range mapping {
		verifyParse(input, expected, t)
	}
}

func TestParseExprEmptyList(t *testing.T) {
	input := "()"
	expected := "()"
	verifyParse(input, expected, t)
}

func TestParseExprSingletonList(t *testing.T) {
	input := "(foo)"
	expected := "(foo)"
	verifyParse(input, expected, t)
}

func TestParseExprList(t *testing.T) {
	input := "(foo  bar    baz)"
	expected := "(foo bar baz)"
	verifyParse(input, expected, t)
}

func TestParseExprNestedList(t *testing.T) {
	input := `(foo
  (bar
    baz))`
	expected := "(foo (bar baz))"
	verifyParse(input, expected, t)
}

func TestParseExprBoolean(t *testing.T) {
	input := "( #t #f )"
	expected := "(#t #f)"
	verifyParse(input, expected, t)
}

func TestParseExprString(t *testing.T) {
	input := `"foo"`
	expected := `"foo"`
	verifyParse(input, expected, t)
}

func TestParseCharacters(t *testing.T) {
	mapping := make(map[string]string)
	mapping["#\\a"] = "#\\a"
	mapping["#\\t"] = "#\\t"
	mapping["#\\newline"] = "#\\newline"
	mapping["#\\space"] = "#\\space"
	mapping["#\\M"] = "#\\M"
	mapping["#\\z"] = "#\\z"
	verifyParseMap(mapping, t)
}

func TestParseQuotes(t *testing.T) {
	mapping := make(map[string]string)
	mapping["(foo 'x)"] = "(foo (quote x))"
	mapping["(foo `x)"] = "(foo (quasiquote x))"
	mapping["(foo ,x)"] = "(foo (unquote x))"
	mapping["(foo ,@x)"] = "(foo (unquote-splicing x))"
	verifyParseMap(mapping, t)
}

func TestParseVector(t *testing.T) {
	input := "#(1 2 3)"
	result, err := parseExpr(input)
	if err != nil {
		msg := fmt.Sprintf("failed to parse expression '%s', %s", input, err)
		t.Errorf(msg)
	} else {
		if slice, ok := result.([]interface{}); ok {
			if len(slice) == 3 {
				if slice[0] != int64(1) && slice[1] != int64(2) && slice[2] != int64(3) {
					t.Errorf("expected 1, 2, 3 in slice, but got %s", slice)
				}
			} else {
				t.Errorf("expected slice of length three but got %d", len(slice))
			}
		} else {
			t.Errorf("expected slice but got %T", result)
		}
	}
}

func TestParseExprNumbers(t *testing.T) {
	mapping := make(map[string]string)
	mapping["1.2345"] = "1.2345"
	mapping[".1"] = "0.1"
	mapping["6e4"] = "60000"
	mapping["12345"] = "12345"
	mapping["2.1"] = "2.1"
	mapping["3."] = "3"
	mapping["7.91e+16"] = "7.91e+16"
	mapping[".000001"] = "1e-06"
	mapping["#b11111111"] = "255"
	mapping["#o777"] = "511"
	mapping["#x4dfCF0"] = "5111024"
	mapping["#d12345"] = "12345"
	mapping["#d#i12345"] = "12345"
	mapping["#d#e12345"] = "12345"
	mapping["#i#d12345"] = "12345"
	mapping["#e#d12345"] = "12345"
	// note that in Go, -0 is the same as 0, so sign will be lost
	mapping["3+4i"] = "3+4i"
	mapping["3-4i"] = "3-4i"
	mapping["3.0+4.0i"] = "3+4i"
	mapping["3.0-4.0i"] = "3-4i"
	mapping["3+i"] = "3+1i"
	mapping["3-i"] = "3-1i"
	mapping["+4i"] = "0+4i"
	mapping["-4i"] = "0-4i"
	mapping["+i"] = "0+1i"
	mapping["-i"] = "0-1i"
	mapping["1/1"] = "1"
	mapping["1/2"] = "0.5"
	mapping["1/3"] = "0.3333333333333333"
	mapping["1/4"] = "0.25"
	mapping["3/4"] = "0.75"
	mapping["6/10"] = "0.6"
	mapping["100/1000"] = "0.1"
	verifyParseMap(mapping, t)
}

func TestExpand(t *testing.T) {
	mapping := make(map[string]string)
	mapping[`(if #t (display "foo"))`] = `(if #t (display "foo") ())`
	mapping[`(quote abc)`] = `(quote abc)`
	mapping[`(set! foo (quote bar))`] = `(set! foo (quote bar))`
	mapping[`(set! foo (if #t (quote bar)))`] = `(set! foo (if #t (quote bar) ()))`
	mapping[`(define (f args) body)`] = `(define f (lambda (args) body))`
	mapping["(define-macro foo (lambda args (if #t (quote bar))))"] =
		"(define-macro foo (lambda args (if #t (quote bar) ())))"
	mapping[`(begin (if #t (display "foo")))`] = `(begin (if #t (display "foo") ()))`
	mapping[`(lambda (x) e1)`] = `(lambda (x) e1)`
	mapping[`(lambda (x) e1 e2)`] = `(lambda (x) (begin e1 e2))`
	mapping[`(foo (if #t (quote bar)))`] = `(foo (if #t (quote bar) ()))`
	mapping["(foo `x)"] = "(foo (quote x))"
	mapping["(foo `,x)"] = "(foo x)"
	// TODO: this is not correct for Scheme
	mapping["(foo `(,@x y))"] = "(foo (append x (cons (quote y) (quote ()))))"
	// TODO: test macro invocation
	verifyExpandMap(mapping, t)
}

func TestExpandErrors(t *testing.T) {
	input := make(map[string]expectedExpandError)
	input["(if)"] = expectedExpandError{"if too many/few arguments", "if requires 3 or 4 args"}
	input["(if bar)"] = expectedExpandError{"if too many/few arguments", "if requires 3 or 4 args"}
	input["(if foo bar baz quux)"] = expectedExpandError{"if too many/few arguments", "if requires 3 or 4 args"}
	input["(set!)"] = expectedExpandError{"set requires 2 arguments", "set requires 2 args"}
	input["(set! foo)"] = expectedExpandError{"set requires 2 arguments", "set requires 2 args"}
	input["(set! (foo) bar)"] = expectedExpandError{"can only set! a symbol", "cannot set non-symbols"}
	input["(set! bar baz quux)"] = expectedExpandError{"set requires 2 arguments", "set requires 2 args"}
	input["(quote)"] = expectedExpandError{"quote requires datum", "quote takes 1 arg"}
	input["(quote foo bar)"] = expectedExpandError{"quote requires datum", "quote takes 1 arg"}
	input["(lambda foo)"] = expectedExpandError{"lambda requires 2+ arguments", "lambda takes 2+ args"}
	input[`(lambda ("foo") bar)`] = expectedExpandError{"lambda arguments must be symbols", "lambda takes symbol args"}
	input[`(lambda "foo" bar)`] = expectedExpandError{"lambda arguments must be a list or a symbol", "lambda takes symbol args"}
	verifyExpandError(t, input)
}
