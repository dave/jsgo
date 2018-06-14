// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file sets up the universe scope and the unsafe package.

package gotypes

// A builtinId is the id of a builtin function.
type BuiltinId int

const (
	// universe scope
	Append BuiltinId = iota
	Cap
	Close
	Complex
	Copy
	Delete
	Imag
	Len
	Make
	New
	Panic
	Print
	Println
	Real
	Recover

	// package unsafe
	Alignof
	Offsetof
	Sizeof

	// testing support
	Assert
	Trace
)

var BuiltinNames = map[string]BuiltinId{
	"append":   Append,
	"cap":      Cap,
	"close":    Close,
	"complex":  Complex,
	"copy":     Copy,
	"delete":   Delete,
	"imag":     Imag,
	"len":      Len,
	"make":     Make,
	"new":      New,
	"panic":    Panic,
	"print":    Print,
	"println":  Println,
	"real":     Real,
	"recover":  Recover,
	"Alignof":  Alignof,
	"Offsetof": Offsetof,
	"Sizeof":   Sizeof,
	"assert":   Assert,
	"trace":    Trace,
}
