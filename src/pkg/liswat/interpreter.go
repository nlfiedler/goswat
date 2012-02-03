//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package liswat

//
// Interpreter for our Scheme-like language, which turns a tree of
// expressions into a final, evaluated result.
//

type Environment struct {
	vars   map[Symbol]interface{} // mapping of variable names to values
	parent *Environment           // environment to delegate to
}

// NewEnvironment constructs an Environment with the given parent to
// provide a fallback for finding variables. The parent may be nil.
func NewEnvironment(parent *Environment) *Environment {
	e := new(Environment)
	e.vars = make(map[Symbol]interface{})
	e.parent = parent
	return e
}

// Find retrieves the value for the given symbol. If it is not found in
// this environment, the parent environment will be consulted. If no
// value is found, nil is returned.
func (e *Environment) Find(sym Symbol) interface{} {
	val, ok := e.vars[sym]
	if !ok {
		if e.parent != nil {
			return e.parent.Find(sym)
		}
		return nil
	}
	return val
}

// Define assigns the given value to the symbol in this environment.
func (e *Environment) Define(sym Symbol, val interface{}) {
	e.vars[sym] = val
}

// Set assigns a value to the given symbol, if and only if that symbol
// has a value already associated with it. If the symbol does not appear
// in this environment, the parent will be consulted.
func (e *Environment) Set(sym Symbol, val interface{}) *LispError {
	_, ok := e.vars[sym]
	if !ok {
		if e.parent != nil {
			return e.parent.Set(sym, val)
		}
		return NewLispError(EVARUNDEF, string(sym)+" undefined")
	} else {
		e.vars[sym] = val
	}
	return nil
}
