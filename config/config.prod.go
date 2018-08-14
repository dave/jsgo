// +build !dev

package config

const (
	DEV = false

	ErrorKind      = "Error"
	CompileKind    = "Compile"
	PackageKind    = "Package"
	DeployKind     = "Deploy"
	ShareKind      = "Share"
	HintsKind      = "Hints"
	WasmDeployKind = "WasmDeploy"
)

var Bucket = map[string]string{
	Src:   "src.jsgo.io",
	Pkg:   "pkg.jsgo.io",
	Index: "jsgo.io",
	Git:   "git.jsgo.io",
}
