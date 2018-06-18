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

var Typ = []*Basic{
	Invalid: {Invalid, 0, "invalid type"},

	Bool:          {Bool, IsBoolean, "bool"},
	Int:           {Int, IsInteger, "int"},
	Int8:          {Int8, IsInteger, "int8"},
	Int16:         {Int16, IsInteger, "int16"},
	Int32:         {Int32, IsInteger, "int32"},
	Int64:         {Int64, IsInteger, "int64"},
	Uint:          {Uint, IsInteger | IsUnsigned, "uint"},
	Uint8:         {Uint8, IsInteger | IsUnsigned, "uint8"},
	Uint16:        {Uint16, IsInteger | IsUnsigned, "uint16"},
	Uint32:        {Uint32, IsInteger | IsUnsigned, "uint32"},
	Uint64:        {Uint64, IsInteger | IsUnsigned, "uint64"},
	Uintptr:       {Uintptr, IsInteger | IsUnsigned, "uintptr"},
	Float32:       {Float32, IsFloat, "float32"},
	Float64:       {Float64, IsFloat, "float64"},
	Complex64:     {Complex64, IsComplex, "complex64"},
	Complex128:    {Complex128, IsComplex, "complex128"},
	String:        {String, IsString, "string"},
	UnsafePointer: {UnsafePointer, 0, "Pointer"},

	UntypedBool:    {UntypedBool, IsBoolean | IsUntyped, "untyped bool"},
	UntypedInt:     {UntypedInt, IsInteger | IsUntyped, "untyped int"},
	UntypedRune:    {UntypedRune, IsInteger | IsUntyped, "untyped rune"},
	UntypedFloat:   {UntypedFloat, IsFloat | IsUntyped, "untyped float"},
	UntypedComplex: {UntypedComplex, IsComplex | IsUntyped, "untyped complex"},
	UntypedString:  {UntypedString, IsString | IsUntyped, "untyped string"},
	UntypedNil:     {UntypedNil, IsUntyped, "untyped nil"},
}
