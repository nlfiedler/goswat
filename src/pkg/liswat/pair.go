//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package liswat

import (
	"bytes"
)

// emptyList represents an empty list and is used by Pair to
// mark the end of a linked list.
var emptyList = "()"

// Pair represents a pair of items, which themselves may be pairs.
type Pair struct {
	first interface{} // the car of the pair
	rest  interface{} // the cdr of the pair
}

// Cons constructs a pair to hold item a and b.
func Cons(a, b interface{}) *Pair {
	return &Pair{a, b}
}

// List constructs a list to hold a and b.
func List(a, b interface{}) *Pair {
	return &Pair{a, &Pair{b, emptyList}}
}

// First returns the first item in the pair.
func (p *Pair) First() interface{} {
	if p != nil {
		return p.first
	}
	return nil
}

// Rest returns the second item in the pair.
func (p *Pair) Rest() interface{} {
	if p != nil {
		return p.rest
	}
	return nil
}

// Second returns the second item in the list, or nil if there is no
// such item.
func (p *Pair) Second() interface{} {
	if p != nil {
		if p.rest == emptyList {
			return nil
		} else if r, ok := p.rest.(*Pair); ok {
			return r.first
		}
		return p.rest
	}
	return nil
}

// Third returns the third item in the list, or nil if there is no such
// item.
func (p *Pair) Third() interface{} {
	if p != nil {
		if r1, ok := p.rest.(*Pair); ok {
			if r1.rest == emptyList {
				return nil
			} else if r2, ok := r1.rest.(*Pair); ok {
				return r2.first
			}
			return r1.rest
		}
	}
	return nil
}

// Reverse returns a new list consisting of the elements in this list in
// reverse order.
func (p *Pair) Reverse() *Pair {
	var result *Pair = nil
	var penultimate *Pair = nil
	for p != nil {
		if result == nil {
			result = Cons(p.first, emptyList)
		} else {
			result = Cons(p.first, result)
			if penultimate == nil {
				penultimate = result
			}
		}
		if p.rest == emptyList {
			p = nil
		} else if r, ok := p.rest.(*Pair); ok {
			p = r
		} else {
			result = Cons(p.rest, result)
			p = nil
		}
	}
	if penultimate != nil {
		// tighten up the end of the list
		if r, ok := penultimate.rest.(*Pair); ok {
			penultimate.rest = r.first
		}
	}
	return result
}

// Len finds the length of the pair, which may be greater than two if
// the pair is part of a list of items.
func (p *Pair) Len() int {
	length := 0
	for p != nil {
		length++
		if p.rest == emptyList {
			p = nil
		} else if r, ok := p.rest.(*Pair); ok {
			p = r
		} else {
			length++
			p = nil
		}
	}
	return length
}

// String returns the string form of the pair.
func (p *Pair) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("(")
	for p != nil {
		stringifyBuffer(p.first, buf)
		if p.rest == emptyList {
			p = nil
		} else if r, ok := p.rest.(*Pair); ok {
			buf.WriteString(" ")
			p = r
		} else {
			buf.WriteString(" . ")
			stringifyBuffer(p.rest, buf)
			p = nil
		}
	}
	buf.WriteString(")")
	return buf.String()
}
