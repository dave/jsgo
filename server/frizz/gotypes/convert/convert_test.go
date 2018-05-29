package convert

import (
	"go/parser"
	"go/token"
	"testing"

	"go/ast"
	"go/types"

	"encoding/json"

	"bytes"
	"fmt"

	"strings"

	"sort"

	"regexp"

	"github.com/dave/jsgo/server/frizz/gotypes"
)

func TestConvertType(t *testing.T) {
	type spec struct {
		code     string
		expected string
	}
	tests := map[string]spec{
		"simple": {
			`type Foo int`,
			`Foo: gotypes.Basic: {"Kind":2,"Info":2,"Name":"int"}`,
		},
		"ignore non-global": {
			`type Foo string
			func f() {
				type Bar string
			}`,
			`Foo: gotypes.Basic: {"Kind":17,"Info":32,"Name":"string"}`,
		},
		"ignore non-exported": {
			`type Foo string
			 type bar string`,
			`Foo: gotypes.Basic: {"Kind":17,"Info":32,"Name":"string"}`,
		},
		"ignore non-exported methods": {
			`type Foo string
			 func (Foo) bar(){}
			 func (Foo) Baz(){}`,
			`Foo: gotypes.Basic: {"Kind":17,"Info":32,"Name":"string"}, methods: [{"Pkg":"foo","Name":"Baz","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":null},"Variadic":false}}]`,
		},
		"ignore non-exported interface methods": {
			`type Foo interface {
				foo()
				Bar()
			}`,
			`Foo: gotypes.Interface: {"Methods":[{"Pkg":"foo","Name":"Bar","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":null},"Variadic":false}}],"Embeddeds":null,"AllMethods":[{"Pkg":"foo","Name":"Bar","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":null},"Variadic":false}}]}`,
		},
		"ignore non-exported interface embeds": {
			`type foo interface{}
			type Bar interface{}
			type Baz interface {
				foo
				Bar
			}`,
			`Bar: gotypes.Interface: {"Methods":null,"Embeddeds":null,"AllMethods":null}
			Baz: gotypes.Interface: {"Methods":null,"Embeddeds":[{"Obj":{"Pkg":"foo","Name":"Bar","Typ":{}},"Type":{"Methods":null,"Embeddeds":null,"AllMethods":null},"Methods":null}],"AllMethods":null}`,
		},
		"include exported alias of non-exported type": {
			`type foo struct {
				Bar string
			}
			type Baz foo`,
			`Baz: gotypes.Struct: {"Fields":[{"Pkg":"foo","Name":"Bar","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":true}],"Tags":[""]}`,
		},
		"include exported pointer to non-exported type": {
			`type foo struct {
				Bar string
			}
			type Baz *foo`,
			`Baz: gotypes.Pointer: {"Elem":{"Obj":{"Pkg":"foo","Name":"foo","Typ":{}},"Type":{"Fields":[{"Pkg":"foo","Name":"Bar","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":true}],"Tags":[""]},"Methods":null}}`,
		},
		"two types": {
			`type Foo int64
			type Bar rune`,
			`Foo: gotypes.Basic: {"Kind":6,"Info":2,"Name":"int64"}
			Bar: gotypes.Basic: {"Kind":5,"Info":2,"Name":"rune"}`,
		},
		"alias": {
			`type Foo int
			type Bar Foo`,
			`Foo: gotypes.Basic: {"Kind":2,"Info":2,"Name":"int"}
			Bar: gotypes.Basic: {"Kind":2,"Info":2,"Name":"int"}`,
		},
		"struct": {
			`type Foo struct {
				Bar string
				baz string
			}`,
			`Foo: gotypes.Struct: {"Fields":[{"Pkg":"foo","Name":"Bar","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":true}],"Tags":[""]}`,
		},
		"array": {
			`type Foo [2]string`,
			`Foo: gotypes.Array: {"Len":2,"Elem":{"Kind":17,"Info":32,"Name":"string"}}`,
		},
		"slice": {
			`type Foo []int`,
			`Foo: gotypes.Slice: {"Elem":{"Kind":2,"Info":2,"Name":"int"}}`,
		},
		"pointer": {
			`type Foo *int`,
			`Foo: gotypes.Pointer: {"Elem":{"Kind":2,"Info":2,"Name":"int"}}`,
		},
		"func type": {
			`type Foo func(int)`,
			`Foo: gotypes.Signature: {"Recv":{"Pkg":"","Name":"","Typ":null,"Anonymous":false,"IsField":false},"Params":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":2,"Info":2,"Name":"int"},"Anonymous":false,"IsField":false}]},"Results":{"Vars":null},"Variadic":false}`,
		},
		"interface": {
			`type Foo interface{
				A() string
				B(int, ...string)
			}`,
			`Foo: gotypes.Interface: {"Methods":[{"Pkg":"foo","Name":"A","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":false}]},"Variadic":false}},{"Pkg":"foo","Name":"B","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":2,"Info":2,"Name":"int"},"Anonymous":false,"IsField":false},{"Pkg":"foo","Name":"","Typ":{"Elem":{"Kind":17,"Info":32,"Name":"string"}},"Anonymous":false,"IsField":false}]},"Results":{"Vars":null},"Variadic":true}}],"Embeddeds":null,"AllMethods":[{"Pkg":"foo","Name":"A","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":false}]},"Variadic":false}},{"Pkg":"foo","Name":"B","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":2,"Info":2,"Name":"int"},"Anonymous":false,"IsField":false},{"Pkg":"foo","Name":"","Typ":{"Elem":{"Kind":17,"Info":32,"Name":"string"}},"Anonymous":false,"IsField":false}]},"Results":{"Vars":null},"Variadic":true}}]}`,
		},
		"interface with embeds": {
			`type Foo interface{
				A() string
			}
			type Bar interface{
				Foo
				B() string
			}`,
			`Foo: gotypes.Interface: {"Methods":[{"Pkg":"foo","Name":"A","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":false}]},"Variadic":false}}],"Embeddeds":null,"AllMethods":[{"Pkg":"foo","Name":"A","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":false}]},"Variadic":false}}]}
			Bar: gotypes.Interface: {"Methods":[{"Pkg":"foo","Name":"B","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":false}]},"Variadic":false}}],"Embeddeds":[{"Obj":{"Pkg":"foo","Name":"Foo","Typ":{}},"Type":{"Methods":[{"Pkg":"foo","Name":"A","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":false}]},"Variadic":false}}],"Embeddeds":null,"AllMethods":[{"Pkg":"foo","Name":"A","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":false}]},"Variadic":false}}]},"Methods":null}],"AllMethods":[{"Pkg":"foo","Name":"A","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{"Obj":{"Pkg":"foo","Name":"Foo","Typ":{}},"Type":{"Methods":[{"Pkg":"foo","Name":"A","Typ":{}}],"Embeddeds":null,"AllMethods":[{"Pkg":"foo","Name":"A","Typ":{}}]},"Methods":null},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":false}]},"Variadic":false}},{"Pkg":"foo","Name":"B","Typ":{"Recv":{"Pkg":"foo","Name":"","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":17,"Info":32,"Name":"string"},"Anonymous":false,"IsField":false}]},"Variadic":false}}]}`,
		},
		"map": {
			`type Foo map[string]int`,
			`Foo: gotypes.Map: {"Key":{"Kind":17,"Info":32,"Name":"string"},"Elem":{"Kind":2,"Info":2,"Name":"int"}}`,
		},
		"chan": {
			`type Foo chan<- int`,
			`Foo: gotypes.Chan: {"Dir":1,"Elem":{"Kind":2,"Info":2,"Name":"int"}}`,
		},
		"methods": {
			`type Foo struct{}
			func (f Foo) Bar() int { return 1 }`,
			`Foo: gotypes.Struct: {"Fields":null,"Tags":null}, methods: [{"Pkg":"foo","Name":"Bar","Typ":{"Recv":{"Pkg":"foo","Name":"f","Typ":{},"Anonymous":false,"IsField":false},"Params":{"Vars":null},"Results":{"Vars":[{"Pkg":"foo","Name":"","Typ":{"Kind":2,"Info":2,"Name":"int"},"Anonymous":false,"IsField":false}]},"Variadic":false}}]`,
		},
		"func": {
			`func Foo() {}`,
			``,
		},
	}
	const single = ""
	if single != "" {
		tests = map[string]spec{single: tests[single]}
	}
	for name, test := range tests {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "foo.go", []byte("package foo\n"+test.code), 0)
		if err != nil {
			t.Fatal(err)
		}
		tc := types.Config{}
		info := &types.Info{
			Types: map[ast.Expr]types.TypeAndValue{},
			Defs:  map[*ast.Ident]types.Object{},
		}

		p, err := tc.Check("foo", fset, []*ast.File{f}, info)
		if err != nil {
			t.Fatal(err)
		}

		var defs []*types.Named
		for _, v := range info.Defs {
			if v == nil {
				continue
			}
			if v.Parent() != p.Scope() {
				continue
			}
			if !v.Exported() {
				continue
			}
			tn, ok := v.(*types.TypeName)
			if !ok {
				continue
			}
			n, ok := tn.Type().(*types.Named)
			if !ok {
				t.Fatalf("%s, got %T", name, v)
			}
			defs = append(defs, n)
		}
		sort.Slice(defs, func(i, j int) bool { return defs[i].Obj().Pos() < defs[j].Obj().Pos() })
		var globals []gotypes.Named
		for _, v := range defs {
			t := Type(v, &[]types.Type{})
			if t == nil {
				continue
			}
			globals = append(globals, t.(gotypes.Named))
		}
		buf := &bytes.Buffer{}
		for _, g := range globals {
			b, err := json.Marshal(g.Type)
			if err != nil {
				t.Fatal(err)
			}
			if len(g.Methods) > 0 {
				mb, err := json.Marshal(g.Methods)
				if err != nil {
					t.Fatal(err)
				}
				fmt.Fprintf(buf, "%s: %T: %s, methods: %s\n", g.Obj.Name, g.Type, string(b), string(mb))
			} else {
				fmt.Fprintf(buf, "%s: %T: %s\n", g.Obj.Name, g.Type, string(b))
			}
		}
		if strings.TrimSpace(buf.String()) != indent.ReplaceAllString(test.expected, "") {
			t.Fatalf("%s, got:\n%s", name, buf.String())
		}
	}
	if single != "" {
		t.Fatal("single test")
	}
}

var indent = regexp.MustCompile(`(?m)^\s*`)
