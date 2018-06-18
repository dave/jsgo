// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gotypes

import (
	"bytes"
	"go/ast"
)

type Object interface {
	Id() Identifier
}

type Identifier struct {
	Path string
	Name string
}

// String returns name if it is exported, otherwise it
// returns the name qualified with the package path.
func (i Identifier) String() string {
	if ast.IsExported(i.Name) {
		return i.Name
	}
	// unexported names need the package path for differentiation
	// (if there's no package, make sure we don't start with '.'
	// as that may change the order of methods between a setup
	// inside a package and outside a package - which breaks some
	// tests)
	path := "_"
	// pkg is nil for objects in Universe scope and possibly types
	// introduced via Eval (see also comment in object.sameId)
	if i.Path != "" {
		path = i.Path
	}
	return path + "." + i.Name
}

func (i Identifier) Exported() bool {
	return ast.IsExported(i.Name)
}

// An object implements the common parts of an Object.
type Obj struct {
	Identifier
	Type Type
}

func (obj Obj) Id() Identifier {
	return obj.Identifier
}

// A PkgName represents an imported Go package.
// PkgNames don't have a type.
type PkgName struct {
	Obj
	Imported string // Imported returns the package that was imported. It is distinct from Pkg(), which is the package containing the import statement.
}

// A Const represents a declared constant.
type Const struct {
	Obj
	Kind ConstKind
}

// Kind specifies the kind of value represented by a Value.
type ConstKind int

const (
	UnknownConst ConstKind = iota
	BoolConst
	StringConst
	IntConst
	FloatConst
	ComplexConst
)

// A TypeName represents a name for a (named or alias) type.
type TypeName struct {
	Obj
}

// A Variable represents a declared variable (including function parameters and results, and struct fields).
type Var struct {
	Obj
	Anonymous bool // Anonymous reports whether the variable is an anonymous field. If set, the variable is an anonymous struct field, and name is the type name
	IsField   bool // IsField reports whether the variable is a struct field.
}

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

// A Label represents a declared label.
// Labels don't have a type.
type Label struct {
	Obj
}

// A Builtin represents a built-in function.
// Builtins don't have a valid type.
type Builtin struct {
	Obj
	BuiltinId BuiltinId
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

func (obj *PkgName) String() string  { return obj.Imported }
func (obj *Const) String() string    { return obj.Name }
func (obj *TypeName) String() string { return obj.Path + "." + obj.Name }
func (obj *Var) String() string      { return obj.Name }
func (obj *Func) String() string     { return obj.FullName() }
func (obj *Label) String() string    { return obj.Name }
func (obj *Builtin) String() string  { return obj.Name }
func (obj *Nil) String() string      { return obj.Name }

func writeFuncName(buf *bytes.Buffer, f Func, qf Qualifier) {
	if f.Type != nil {
		sig := f.Type.(*Signature)
		if recv := sig.Recv; recv != nil {
			buf.WriteByte('(')
			if _, ok := recv.Type.(*Interface); ok {
				// gcimporter creates abstract methods of
				// named interfaces using the interface type
				// (not the named type) as the receiver.
				// Don't print it in full.
				buf.WriteString("interface")
			} else {
				WriteType(buf, recv.Type, qf)
			}
			buf.WriteByte(')')
			buf.WriteByte('.')
		} else if f.Path != "" {
			writePackage(buf, f.Path, qf)
		}
	}
	buf.WriteString(f.Name)
}
