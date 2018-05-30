// +build !dev,!local

package config

const (
	LOCAL = false
)

var Host = map[string]string{
	Jsgo:  "compile.jsgo.io",
	Play:  "play.jsgo.io",
	Frizz: "frizz.io",
	Src:   "src.jsgo.io",
	Pkg:   "pkg.jsgo.io",
	Index: "jsgo.io",
}

var Protocol = map[string]string{
	Jsgo:  "https",
	Play:  "https",
	Frizz: "https",
	Src:   "https",
	Pkg:   "https",
	Index: "https",
}
