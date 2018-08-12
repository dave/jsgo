// +build dev,!local

package config

const (
	LOCAL = false
)

var Host = map[string]string{
	Play:  "localhost:8080",
	Jsgo:  "localhost:8081",
	Frizz: "localhost:8082",
	Wasm:  "localhost:8083",
	Src:   "dev-src.jsgo.io",
	Pkg:   "dev-pkg.jsgo.io",
	Index: "dev-index.jsgo.io",
}

var Protocol = map[string]string{
	Jsgo:  "http",
	Play:  "http",
	Frizz: "http",
	Wasm:  "http",
	Src:   "https",
	Pkg:   "https",
	Index: "https",
}
