// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gotypes

// A Type represents a type of Go.
// All types implement the Type interface.
type Type interface {
	// Underlying returns the underlying type of a type.
	Underlying() Type

	// String returns a string representation of a type.
	String() string
}

// When a circular reference is detected, this is used as a placeholder
type Circular struct{}

func (c Circular) Underlying() Type { return nil }
func (c Circular) String() string   { return "circular reference" }

// BasicKind describes the kind of basic type.
type BasicKind int

const (
	Invalid BasicKind = iota // type is invalid

	// predeclared types
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	String
	UnsafePointer

	// types for untyped values
	UntypedBool
	UntypedInt
	UntypedRune
	UntypedFloat
	UntypedComplex
	UntypedString
	UntypedNil

	// aliases
	Byte = Uint8
	Rune = Int32
)

// BasicInfo is a set of flags describing properties of a basic type.
type BasicInfo int

// Properties of basic types.
const (
	IsBoolean BasicInfo = 1 << iota
	IsInteger
	IsUnsigned
	IsFloat
	IsComplex
	IsString
	IsUntyped

	IsOrdered   = IsInteger | IsFloat | IsString
	IsNumeric   = IsInteger | IsFloat | IsComplex
	IsConstType = IsBoolean | IsNumeric | IsString
)

// A Basic represents a basic type.
type Basic struct {
	Kind BasicKind // Kind returns the kind of basic type b.
	Info BasicInfo // Info returns information about properties of basic type b.
	Name string    // Name returns the name of basic type b.
}

// An Array represents an array type.
type Array struct {
	Len  int64 // Len returns the length of array a.
	Elem Type  // Elem returns element type of array a.
}

// A Slice represents a slice type.
type Slice struct {
	Elem Type // Elem returns the element type of slice s.
}

// A Struct represents a struct type.
type Struct struct {
	Fields []Var
	Tags   []string // field tags; nil if there are no tags
}

// NumFields returns the number of fields in the struct (including blank and anonymous fields).
func (s Struct) NumFields() int { return len(s.Fields) }

// Field returns the i'th field for 0 <= i < NumFields().
func (s Struct) Field(i int) Var { return s.Fields[i] }

// Tag returns the i'th field tag for 0 <= i < NumFields().
func (s Struct) Tag(i int) string {
	if i < len(s.Tags) {
		return s.Tags[i]
	}
	return ""
}

// A Pointer represents a pointer type.
type Pointer struct {
	Elem Type // Elem returns the element type for the given pointer p.
}

// A Tuple represents an ordered list of variables; a nil *Tuple is a valid (empty) tuple.
// Tuples are used as components of signatures and to represent the type of multiple
// assignments; they are not first class types of Go.
type Tuple struct {
	Vars []Var
}

// Len returns the number variables of tuple t.
func (t Tuple) Len() int {
	return len(t.Vars)
}

// At returns the i'th variable of tuple t.
func (t Tuple) At(i int) Var { return t.Vars[i] }

// A Signature represents a (non-builtin) function or method type.
// The receiver is ignored when comparing signatures for identity.
type Signature struct {
	// Recv returns the receiver of signature s (if a method), or nil if a
	// function. It is ignored when comparing signatures for identity.
	//
	// For an abstract method, Recv returns the enclosing interface either
	// as a *Named or an *Interface. Due to embedding, an interface may
	// contain methods whose receiver type is a different interface.
	Recv     Var
	Params   Tuple // Params returns the parameters of signature s, or nil.
	Results  Tuple // Results returns the results of signature s, or nil.
	Variadic bool  // Variadic reports whether the signature s is variadic.
}

// An Interface represents an interface type.
type Interface struct {
	Methods   []Func  // ordered list of explicitly declared methods
	Embeddeds []Named // ordered list of explicitly embedded types

	AllMethods []Func // ordered list of methods declared with or embedded in this interface (TODO(gri): replace with mset)
}

// EmptyInterface represents the empty (completed) interface
var EmptyInterface = Interface{AllMethods: MarkComplete}

// MarkComplete is used to mark an empty interface as completely
// set up by setting the allMethods field to a non-nil empty slice.
var MarkComplete = make([]Func, 0)

// NumExplicitMethods returns the number of explicitly declared methods of interface t.
func (t Interface) NumExplicitMethods() int { return len(t.Methods) }

// ExplicitMethod returns the i'th explicitly declared method of interface t for 0 <= i < t.NumExplicitMethods().
// The methods are ordered by their unique Id.
func (t Interface) ExplicitMethod(i int) Func { return t.Methods[i] }

// NumEmbeddeds returns the number of embedded types in interface t.
func (t Interface) NumEmbeddeds() int { return len(t.Embeddeds) }

// Embedded returns the i'th embedded type of interface t for 0 <= i < t.NumEmbeddeds().
// The types are ordered by the corresponding TypeName's unique Id.
func (t Interface) Embedded(i int) Named { return t.Embeddeds[i] }

// NumMethods returns the total number of methods of interface t.
func (t Interface) NumMethods() int { return len(t.AllMethods) }

// Method returns the i'th method of interface t for 0 <= i < t.NumMethods().
// The methods are ordered by their unique Id.
func (t Interface) Method(i int) Func { return t.AllMethods[i] }

// Empty returns true if t is the empty interface.
func (t Interface) Empty() bool { return len(t.AllMethods) == 0 }

// A Map represents a map type.
type Map struct {
	Key  Type // Key returns the key type of map m.
	Elem Type // Elem returns the element type of map m.
}

// A Chan represents a channel type.
type Chan struct {
	Dir  ChanDir // Dir returns the direction of channel c.
	Elem Type    // Elem returns the element type of channel c.
}

// A ChanDir value indicates a channel direction.
type ChanDir int

// The direction of a channel is indicated by one of these constants.
const (
	SendRecv ChanDir = iota
	SendOnly
	RecvOnly
)

// A Named represents a named type.
type Named struct {
	Obj     TypeName // corresponding declared object
	Type    Type     // possibly a Named during setup; never a Named once set up completely
	Methods []Func   // methods declared for this type (not the method set of this type)
}

// NumMethods returns the number of explicit methods whose receiver is named type t.
func (t Named) NumMethods() int { return len(t.Methods) }

// Method returns the i'th method of named type t for 0 <= i < t.NumMethods().
func (t Named) Method(i int) Func { return t.Methods[i] }

// Implementations for Type methods.

func (t Basic) Underlying() Type     { return t }
func (t Array) Underlying() Type     { return t }
func (t Slice) Underlying() Type     { return t }
func (t Struct) Underlying() Type    { return t }
func (t Pointer) Underlying() Type   { return t }
func (t Tuple) Underlying() Type     { return t }
func (t Signature) Underlying() Type { return t }
func (t Interface) Underlying() Type { return t }
func (t Map) Underlying() Type       { return t }
func (t Chan) Underlying() Type      { return t }
func (t Named) Underlying() Type     { return t.Type }

func (t Basic) String() string     { return TypeString(t, nil) }
func (t Array) String() string     { return TypeString(t, nil) }
func (t Slice) String() string     { return TypeString(t, nil) }
func (t Struct) String() string    { return TypeString(t, nil) }
func (t Pointer) String() string   { return TypeString(t, nil) }
func (t Tuple) String() string     { return TypeString(t, nil) }
func (t Signature) String() string { return TypeString(t, nil) }
func (t Interface) String() string { return TypeString(t, nil) }
func (t Map) String() string       { return TypeString(t, nil) }
func (t Chan) String() string      { return TypeString(t, nil) }
func (t Named) String() string     { return TypeString(t, nil) }
