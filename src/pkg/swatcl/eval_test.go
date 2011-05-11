//
// Copyright 2011 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"testing"
)

func TestEvalBoolean(t *testing.T) {
	badbools := [...]string{"foo", "1.0", "sure", "yesarooney", "no, sir"}
	for bad := range badbools {
		_, err := evalBoolean(badbools[bad])
		if err == nil || err.Errno != EBADBOOL {
			t.Error("expected bad boolean error")
		}
	}
	tests := make(map[string]bool)
	tests["FaLse"] = false
	tests["tRUE"] = true
	tests["No"] = false
	tests["yeS"] = true
	tests["0"] = false
	tests["1"] = true
	tests["10"] = true
	for k, v := range tests {
		b, err := evalBoolean(k)
		if err != nil {
			t.Errorf("unexpected error in evalBoolean: %s", err)
		}
		if b != v {
			t.Errorf("expected %t for %s", v, k)
		}
	}
}

// evalStrAndCompare invokes evalString on each of the map keys and
// compares the result to the corresponding map value.
func evalStrAndCompare(values map[string]string, t *testing.T) {
	for k, v := range values {
		r, e := evalString(k)
		if e != nil {
			t.Errorf("evaluation of '%s' failed: %s", k, e)
		}
		if r != v {
			t.Errorf("evaluation of '%s' resulted in '%s'", k, r)
		}
	}
}

func TestEvalString(t *testing.T) {
	values := make(map[string]string)
	values["abc"] = "abc"
	values["abc\\adef"] = "abc\adef"
	values["abc\\bdef"] = "abc\bdef"
	values["abc\\fdef"] = "abc\fdef"
	values["abc\\ndef"] = "abc\ndef"
	values["abc\\rdef"] = "abc\rdef"
	values["abc\\tdef"] = "abc\tdef"
	values["abc\\vdef"] = "abc\vdef"
	values["foo\\u005cbar"] = "foo\\bar"
	values["foo\\x5cbar"] = "foo\\bar"
	values["foo\\043bar"] = "foo#bar"
	evalStrAndCompare(values, t)
}
