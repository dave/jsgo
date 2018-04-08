// +build dev

package config

const (
	DEV = true

	SrcBucket   = "dev-src.jsgo.io"
	PkgBucket   = "dev-pkg.jsgo.io"
	IndexBucket = "dev-index.jsgo.io"
	GitBucket   = "dev-git.jsgo.io"

	SrcHost   = "dev-src.jsgo.io"
	PkgHost   = "dev-pkg.jsgo.io"
	IndexHost = "dev-index.jsgo.io"

	ErrorKind   = "ErrorDev"
	CompileKind = "CompileDev"
	PackageKind = "PackageDev"
	DeployKind  = "DeployDev"
	ShareKind   = "ShareDev"
	HintsKind   = "HintsDev"
)
