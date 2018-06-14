// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements printing of types.

package gotypes

import (
	"bytes"
	"fmt"
)

// A Qualifier controls how named package-level objects are printed in
// calls to TypeString, ObjectString, and SelectionString.
//
// These three formatting routines call the Qualifier for each
// package-level object O, and if the Qualifier returns a non-empty
// string p, the object is printed in the form p.O.
// If it returns an empty string, only the object name O is printed.
//
// Using a nil Qualifier is equivalent to using (*Package).Path: the
// object is qualified by the import path, e.g., "encoding/json.Marshal".
//
type Qualifier func(string) string

// RelativeTo(pkg) returns a Qualifier that fully qualifies members of
// all packages other than pkg.
func RelativeTo(pkg string) Qualifier {
	if pkg == "" {
		return nil
	}
	return func(other string) string {
		if pkg == other {
			return "" // same package; unqualified
		}
		return other
	}
}

// TypeString returns the string representation of typ.
// The Qualifier controls the printing of
// package-level objects, and may be nil.
func TypeString(typ Type, qf Qualifier) string {
	var buf bytes.Buffer
	WriteType(&buf, typ, qf)
	return buf.String()
}

// WriteType writes the string representation of typ to buf.
// The Qualifier controls the printing of
// package-level objects, and may be nil.
func WriteType(buf *bytes.Buffer, typ Type, qf Qualifier) {
	writeType(buf, typ, qf, make([]Type, 0, 8))
}

func writeType(buf *bytes.Buffer, typ Type, qf Qualifier, visited []Type) {
	// Theoretically, this is a quadratic lookup algorithm, but in
	// practice deeply nested composite types with unnamed component
	// types are uncommon. This code is likely more efficient than
	// using a map.
	for _, t := range visited {
		if t == typ {
			fmt.Fprintf(buf, "○%T", typ) // cycle to typ
			return
		}
	}
	visited = append(visited, typ)

	switch t := typ.(type) {
	case nil:
		buf.WriteString("<nil>")

	case Basic:
		if t.Kind == UnsafePointer {
			buf.WriteString("unsafe.")
		}
		buf.WriteString(t.Name)

	case Array:
		fmt.Fprintf(buf, "[%d]", t.Len)
		writeType(buf, t.Elem, qf, visited)

	case Slice:
		buf.WriteString("[]")
		writeType(buf, t.Elem, qf, visited)

	case Struct:
		buf.WriteString("struct{")
		for i, f := range t.Fields {
			if i > 0 {
				buf.WriteString("; ")
			}
			if !f.Anonymous {
				buf.WriteString(f.Name)
				buf.WriteByte(' ')
			}
			writeType(buf, f.Type, qf, visited)
			if tag := t.Tag(i); tag != "" {
				fmt.Fprintf(buf, " %q", tag)
			}
		}
		buf.WriteByte('}')

	case Pointer:
		buf.WriteByte('*')
		writeType(buf, t.Elem, qf, visited)

	case Tuple:
		writeTuple(buf, t, false, qf, visited)

	case Signature:
		buf.WriteString("func")
		writeSignature(buf, t, qf, visited)

	case Interface:
		// We write the source-level methods and embedded types rather
		// than the actual method set since resolved method signatures
		// may have non-printable cycles if parameters have anonymous
		// interface types that (directly or indirectly) embed the
		// current interface. For instance, consider the result type
		// of m:
		//
		//     type T interface{
		//         m() interface{ T }
		//     }
		//
		buf.WriteString("interface{")
		empty := true
		// print explicit interface methods and embedded types
		for i, m := range t.Methods {
			if i > 0 {
				buf.WriteString("; ")
			}
			buf.WriteString(m.Name)
			writeSignature(buf, m.Type.(Signature), qf, visited)
			empty = false
		}
		for i, typ := range t.Embeddeds {
			if i > 0 || len(t.Methods) > 0 {
				buf.WriteString("; ")
			}
			writeType(buf, typ, qf, visited)
			empty = false
		}
		if t.AllMethods == nil || len(t.Methods) > len(t.AllMethods) {
			if !empty {
				buf.WriteByte(' ')
			}
			buf.WriteString("/* incomplete */")
		}
		buf.WriteByte('}')

	case Map:
		buf.WriteString("map[")
		writeType(buf, t.Key, qf, visited)
		buf.WriteByte(']')
		writeType(buf, t.Elem, qf, visited)

	case Chan:
		var s string
		var parens bool
		switch t.Dir {
		case SendRecv:
			s = "chan "
			// chan (<-chan T) requires parentheses
			if c, _ := t.Elem.(Chan); c.Dir == RecvOnly {
				parens = true
			}
		case SendOnly:
			s = "chan<- "
		case RecvOnly:
			s = "<-chan "
		default:
			panic("unreachable")
		}
		buf.WriteString(s)
		if parens {
			buf.WriteByte('(')
		}
		writeType(buf, t.Elem, qf, visited)
		if parens {
			buf.WriteByte(')')
		}

	case Named:
		buf.WriteString(fmt.Sprintf("<named type with %d methods>", len(t.Methods)))

	case Reference:
		if t.Path != "" {
			writePackage(buf, t.Path, qf)
		}
		buf.WriteString(t.Name)

	default:
		// For externally defined implementations of Type.
		buf.WriteString(t.String())
	}
}

func writeTuple(buf *bytes.Buffer, tup Tuple, variadic bool, qf Qualifier, visited []Type) {
	buf.WriteByte('(')
	for i, v := range tup.Vars {
		if i > 0 {
			buf.WriteString(", ")
		}
		if v.Name != "" {
			buf.WriteString(v.Name)
			buf.WriteByte(' ')
		}
		typ := v.Type
		if variadic && i == len(tup.Vars)-1 {
			if s, ok := typ.(Slice); ok {
				buf.WriteString("...")
				typ = s.Elem
			} else {
				// special case:
				// append(s, "foo"...) leads to signature func([]byte, string...)
				if t, ok := typ.Underlying().(Basic); !ok || t.Kind != String {
					panic("internal error: string type expected")
				}
				writeType(buf, typ, qf, visited)
				buf.WriteString("...")
				continue
			}
		}
		writeType(buf, typ, qf, visited)
	}
	buf.WriteByte(')')
}

// WriteSignature writes the representation of the signature sig to buf,
// without a leading "func" keyword.
// The Qualifier controls the printing of
// package-level objects, and may be nil.
func WriteSignature(buf *bytes.Buffer, sig Signature, qf Qualifier) {
	writeSignature(buf, sig, qf, make([]Type, 0, 8))
}

func writeSignature(buf *bytes.Buffer, sig Signature, qf Qualifier, visited []Type) {
	writeTuple(buf, sig.Params, sig.Variadic, qf, visited)

	n := sig.Results.Len()
	if n == 0 {
		// no result
		return
	}

	buf.WriteByte(' ')
	if n == 1 && sig.Results.Vars[0].Name == "" {
		// single unnamed result
		writeType(buf, sig.Results.Vars[0].Type, qf, visited)
		return
	}

	// multiple or named result(s)
	writeTuple(buf, sig.Results, false, qf, visited)
}
