//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package swatcl

import (
	"testing"
)

func TestTclResultNil(t *testing.T) {
	var r *tclResult = nil
	if r.ErrorCode() != EBADSTATE {
		t.Error("nil TclResult.ErrorCode() should return EBADSTATE")
	}
	if r.ErrorMessage() != "" {
		t.Error("nil TclResult.ErrorMessage() should return ''")
	}
	if r.Ok() {
		t.Error("nil TclResult.Ok() should return false")
	}
	if r.Result() != "<nil>" {
		t.Error("nil TclResult.Result() should return '<nil>'")
	}
	if r.ReturnCode() != returnError {
		t.Error("nil TclResult.ReturnCode() should return 'returnError'")
	}
	if r.String() != "" {
		t.Error("nil TclResult.String() should return ''")
	}
}

func TestTclResultOk(t *testing.T) {
	r := newTclResult(EOK, "", returnOk, "foo")
	if r.ErrorCode() != EOK {
		t.Error("EOK TclResult.Error() should return EOK")
	}
	if r.ErrorMessage() != "" {
		t.Error("EOK TclResult.ErrorMessage() should return ''")
	}
	if !r.Ok() {
		t.Error("EOK TclResult.Ok() should return true")
	}
	if r.ReturnCode() != returnOk {
		t.Error("returnOk TclResult.ReturnCode() should return 'returnOk'")
	}
	if r.Result() != "foo" {
		t.Error("'foo' TclResult.Result() should return 'foo'")
	}
}
