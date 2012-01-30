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

// TODO: write Environment struct to maintain a collection of
//       key/value pairs for proc parameters and global mappings,
//       in addition to a parent Environment to fall back on
