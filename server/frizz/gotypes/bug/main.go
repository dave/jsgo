package main

import (
	"github.com/gopherjs/gopherjs/js"
)

// Demonstrates a bug in GopherJS, fixed by 69f16807272a9f0023ce6503c51a31725b23321b
// TODO: Do an issue and PR for this.
func main() {
	type Foo struct{}
	js.Global.Get("console").Call("log", Foo{})
}
