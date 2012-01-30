//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package liswat

import (
	"testing"
)

func TestPairNil(t *testing.T) {
	var p *Pair
	if p.Len() != 0 {
		t.Errorf("nil Pair.Len() should be 0")
	}
	if p.First() != nil {
		t.Errorf("nil Pair.First() should be nil")
	}
	if p.Second() != nil {
		t.Errorf("nil Pair.Second() should be nil")
	}
	if p.Third() != nil {
		t.Errorf("nil Pair.Third() should be nil")
	}
	if p.String() != "()" {
		t.Errorf("nil Pair.String() should be ()")
	}
	if p.Reverse() != nil {
		t.Errorf("nil Pair.Reverse() should be nil")
	}
}

func TestList(t *testing.T) {
	foo := Symbol("foo")
	bar := Symbol("bar")
	p := List(foo, bar)
	if p.Len() != 2 {
		t.Errorf("expected 2, but got %d", p.Len())
	}
	if p.First() != foo {
		t.Errorf("expected 'foo', but got '%s'", p.First())
	}
	if p.Second() != bar {
		t.Errorf("expected 'bar', but got '%s'", p.Second())
	}
	if p.Third() != nil {
		t.Errorf("expected nil, but got '%s'", p.Third())
	}
	if p.String() != "(foo bar)" {
		t.Errorf("expected '(foo bar)' but got '%s'", p.String())
	}
	p = p.Reverse()
	if p.Len() != 2 {
		t.Errorf("expected 2, but got %d", p.Len())
	}
	if p.First() != bar {
		t.Errorf("expected 'bar', but got '%s'", p.First())
	}
	if p.Second() != foo {
		t.Errorf("expected 'foo', but got '%s'", p.Second())
	}
	if p.Third() != nil {
		t.Errorf("expected nil, but got '%s'", p.Third())
	}
	if p.String() != "(bar foo)" {
		t.Errorf("expected (bar foo) but got '%s'", p.String())
	}
}

func TestCons(t *testing.T) {
	foo := Symbol("foo")
	bar := Symbol("bar")
	baz := Symbol("baz")
	qux := Symbol("qux")
	p := List(baz, qux)
	p = Cons(bar, p)
	p = Cons(foo, p)
	if p.Len() != 4 {
		t.Errorf("expected 4, but got %d", p.Len())
	}
	if p.First() != foo {
		t.Errorf("expected 'foo', but got '%s'", p.First())
	}
	if p.Second() != bar {
		t.Errorf("expected 'bar', but got '%s'", p.Second())
	}
	if p.Third() != baz {
		t.Errorf("expected 'baz', but got '%s'", p.Third())
	}
	if p.String() != "(foo bar baz qux)" {
		t.Errorf("expected (foo bar baz qux) but got '%s'", p.String())
	}
	p = p.Reverse()
	if p.Len() != 4 {
		t.Errorf("expected 4, but got %d", p.Len())
	}
	if p.First() != qux {
		t.Errorf("expected 'qux', but got '%s'", p.First())
	}
	if p.Second() != baz {
		t.Errorf("expected 'baz', but got '%s'", p.Second())
	}
	if p.Third() != bar {
		t.Errorf("expected 'bar', but got '%s'", p.Third())
	}
	if p.String() != "(qux baz bar foo)" {
		t.Errorf("expected (qux baz bar foo) but got '%s'", p.String())
	}
}
