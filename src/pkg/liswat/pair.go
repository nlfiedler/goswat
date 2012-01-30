//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package liswat

import (
	"bytes"
)

// Pair represents a pair of items, which themselves may be instances of
// Pair. A Pair will contain one item and may contain a reference to
// another Pair, forming a list. nil is used to mark the end of the
// list.
type Pair struct {
	first interface{}  // the car of the pair
	rest  *Pair        // the cdr of the pair
}

// Cons constructs a pair to hold a and b.
func Cons(a interface{}, b *Pair) *Pair {
	return &Pair{a, b}
}

// List constructs a list to hold a and b in separate pairs.
func List(a, b interface{}) *Pair {
	return &Pair{a, &Pair{b, nil}}
}

// First returns the first item in the pair.
func (p *Pair) First() interface{} {
	if p != nil {
		return p.first
	}
	return nil
}

// Rest returns the rest second item in the pair, which may be another
// instance of Pair.
func (p *Pair) Rest() *Pair {
	if p != nil {
		return p.rest
	}
	return nil
}

// Second returns the second item in the list, or nil if there is no
// such item.
func (p *Pair) Second() interface{} {
	if p != nil && p.rest != nil {
		return p.rest.first
	}
	return nil
}

// Third returns the third item in the list, or nil if there is no such
// item.
func (p *Pair) Third() interface{} {
	if p != nil && p.rest != nil && p.rest.rest != nil {
		return p.rest.rest.first
	}
	return nil
}

// Reverse returns a new list consisting of the elements in this list in
// reverse order.
func (p *Pair) Reverse() *Pair {
	var result *Pair
	for p != nil {
		result = Cons(p.first, result)
		p = p.rest
	}
	return result
}

// Len finds the length of the pair, which may be greater than two if
// the pair is part of a list of items.
func (p *Pair) Len() int {
	length := 0
	for p != nil {
		length++
		p = p.rest
	}
	return length
}

// String returns the string form of the pair.
func (p *Pair) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("(")
	if p != nil {
		stringifyBuffer(p.first, buf)
		p = p.rest
		for p != nil {
      			buf.WriteString(" ")
      			stringifyBuffer(p.first, buf)
			p = p.rest
		}
	}
	buf.WriteString(")")
	return buf.String()
 }
