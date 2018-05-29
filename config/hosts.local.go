// +build local

package config

const (
	LOCAL = true
)

var Host = map[Site]string{
	Play:  "localhost:8080",
	Jsgo:  "localhost:8081",
	Frizz: "localhost:8082",
	Src:   "localhost:8091",
	Pkg:   "localhost:8092",
	Index: "localhost:8093",
}

var Protocol = map[Site]string{
	Jsgo:  "http",
	Play:  "http",
	Frizz: "http",
	Src:   "http",
	Pkg:   "http",
	Index: "http",
}
