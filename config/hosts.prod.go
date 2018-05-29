// +build !dev,!local

package config

const (
	LOCAL = false
)

var Host = map[Site]string{
	Jsgo:  "compile.jsgo.io",
	Play:  "play.jsgo.io",
	Frizz: "frizz.io",
	Src:   "src.jsgo.io",
	Pkg:   "pkg.jsgo.io",
	Index: "jsgo.io",
}

var Protocol = map[Site]string{
	Jsgo:  "https",
	Play:  "https",
	Frizz: "https",
	Src:   "https",
	Pkg:   "https",
	Index: "https",
}
