package main

// env GOARCH=wasm GOOS=js go1.11beta3 build -o="foo.wasm" github.com/dave/jsgo/cmd/helloworld
func main() {
	println("Hello, world! 3")
}
