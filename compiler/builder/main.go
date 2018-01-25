package main

import (
	"github.com/dave/jsgo/compiler"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

func main() {
	if err := Do(); err != nil {
		panic(err)
	}
}

func Do() error {
	fs := memfs.New()
	c := compiler.New(fs)
	if err := c.CompileStdLib(); err != nil {
		return err
	}
	return nil
}
