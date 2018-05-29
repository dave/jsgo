// +build dev

package config

const (
	DEV = true

	ErrorKind   = "ErrorDev"
	CompileKind = "CompileDev"
	PackageKind = "PackageDev"
	DeployKind  = "DeployDev"
	ShareKind   = "ShareDev"
	HintsKind   = "HintsDev"
)

var Bucket = map[Site]string{
	Src:   "dev-src.jsgo.io",
	Pkg:   "dev-pkg.jsgo.io",
	Index: "dev-index.jsgo.io",
	Git:   "dev-git.jsgo.io",
}
