// +build !dev

package config

const (
	DEV = false

	SrcBucket   = "src.jsgo.io"
	PkgBucket   = "pkg.jsgo.io"
	IndexBucket = "jsgo.io"
	GitBucket   = "git.jsgo.io"

	SrcHost   = "src.jsgo.io"
	PkgHost   = "pkg.jsgo.io"
	IndexHost = "jsgo.io"

	ErrorKind   = "Error"
	CompileKind = "Compile"
	PackageKind = "Package"
	DeployKind  = "Deploy"
	ShareKind   = "Share"
	HintsKind   = "Hints"
)
