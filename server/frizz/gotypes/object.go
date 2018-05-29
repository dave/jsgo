// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gotypes

import (
	"bytes"
	"go/ast"
	"go/constant"
)

// Id returns name if it is exported, otherwise it
// returns the name qualified with the package path.
func Id(pkg string, name string) string {
	if ast.IsExported(name) {
		return name
	}
	// unexported names need the package path for differentiation
	// (if there's no package, make sure we don't start with '.'
	// as that may change the order of methods between a setup
	// inside a package and outside a package - which breaks some
	// tests)
	path := "_"
	// pkg is nil for objects in Universe scope and possibly types
	// introduced via Eval (see also comment in object.sameId)
	if pkg != "" {
		path = pkg
	}
	return path + "." + name
}

// An object implements the common parts of an Object.
type Obj struct {
	Pkg  string
	Name string
	Typ  Type
}

func (obj Obj) Exported() bool { return ast.IsExported(obj.Name) }
func (obj Obj) Id() string     { return Id(obj.Pkg, obj.Name) }
func (obj Obj) String() string { panic("abstract") }

// A PkgName represents an imported Go package.
// PkgNames don't have a type.
type PkgName struct {
	Obj
	Imported string // Imported returns the package that was imported. It is distinct from Pkg(), which is the package containing the import statement.
}

// A Const represents a declared constant.
type Const struct {
	Obj
	Val constant.Value
}

func (Const) isDependency() {} // a constant may be a dependency of an initialization expression

// A TypeName represents a name for a (named or alias) type.
type TypeName struct {
	Obj
}

// IsAlias reports whether obj is an alias name for a type.
func (obj TypeName) IsAlias() bool {
	switch t := obj.Typ.(type) {
	case nil:
		return false
	case Basic:
		// unsafe.Pointer is not an alias.
		if obj.Pkg == "unsafe" {
			return false
		}
		// Any user-defined type name for a basic type is an alias for a
		// basic type (because basic types are pre-declared in the Universe
		// scope, outside any package scope), and so is any type name with
		// a different name than the name of the basic type it refers to.
		// Additionally, we need to look for "byte" and "rune" because they
		// are aliases but have the same names (for better error messages).
		return obj.Pkg != "" || t.Name != obj.Name || t.Name == "byte" || t.Name == "rune"
	case Named:
		return obj != t.Obj
	default:
		return true
	}
}

// A Variable represents a declared variable (including function parameters and results, and struct fields).
type Var struct {
	Obj
	Anonymous bool // Anonymous reports whether the variable is an anonymous field. If set, the variable is an anonymous struct field, and name is the type name
	IsField   bool // IsField reports whether the variable is a struct field.
}

func (Var) isDependency() {} // a variable may be a dependency of an initialization expression

// A Func represents a declared function, concrete method, or abstract
// (interface) method. Its Type() is always a Signature.
// An abstract method may belong to many interfaces due to embedding.
type Func struct {
	Obj
}

// FullName returns the package- or receiver-type-qualified name of
// function or method obj.
func (obj Func) FullName() string {
	var buf bytes.Buffer
	writeFuncName(&buf, obj, nil)
	return buf.String()
}

func (Func) isDependency() {} // a function may be a dependency of an initialization expression

// A Label represents a declared label.
// Labels don't have a type.
type Label struct {
	Obj
}

// A Builtin represents a built-in function.
// Builtins don't have a valid type.
type Builtin struct {
	Obj
	Id BuiltinId
}

// Nil represents the predeclared value nil.
type Nil struct {
	Obj
}

func writePackage(buf *bytes.Buffer, pkg string, qf Qualifier) {
	if pkg == "" {
		return
	}
	var s string
	if qf != nil {
		s = qf(pkg)
	} else {
		s = pkg
	}
	if s != "" {
		buf.WriteString(s)
		buf.WriteByte('.')
	}
}

func (obj PkgName) String() string  { return obj.Pkg }
func (obj Const) String() string    { return obj.Name }
func (obj TypeName) String() string { return obj.Pkg + "." + obj.Name }
func (obj Var) String() string      { return obj.Name }
func (obj Func) String() string     { return obj.FullName() }
func (obj Label) String() string    { return obj.Name }
func (obj Builtin) String() string  { return obj.Name }
func (obj Nil) String() string      { return obj.Name }

func writeFuncName(buf *bytes.Buffer, f Func, qf Qualifier) {
	if f.Typ != nil {
		sig := f.Typ.(Signature)
		if recv := sig.Recv; recv != (Var{}) {
			buf.WriteByte('(')
			if _, ok := recv.Typ.(Interface); ok {
				// gcimporter creates abstract methods of
				// named interfaces using the interface type
				// (not the named type) as the receiver.
				// Don't print it in full.
				buf.WriteString("interface")
			} else {
				WriteType(buf, recv.Typ, qf)
			}
			buf.WriteByte(')')
			buf.WriteByte('.')
		} else if f.Pkg != "" {
			writePackage(buf, f.Pkg, qf)
		}
	}
	buf.WriteString(f.Name)
}
