package config

import (
	"time"
)

const (
	Jsgo  = "jsgo"
	Wasm  = "wasm"
	Play  = "play"
	Frizz = "frizz"
	Index = "index"
	Pkg   = "pkg"
	Src   = "src"
	Git   = "git"
)

const (
	LocalFileserverTempDir = "/Users/dave/.jsgo-local"

	// ProjectId is the ID of the GCS project
	ProjectID = "jsgo-192815"

	// CompileHost is the domain of the compile server
	CompileHost = "compile.jsgo.io"

	// MaxConcurrentCompiles is the maximum number of concurrent compile jobs per server
	MaxConcurrentCompiles = 2

	// MaxQueue is the maximum queue length waiting for compile. After this an error is returned.
	MaxQueue = 100

	AssetsFilename = "assets.zip"

	// WriteTimeout is the timeout when serving static files
	WriteTimeout = time.Second * 2

	// CompileTimeout is the timeout when compiling a package.
	RequestTimeout = time.Second * 300

	// PageTimeout is the timeout when generating the compile page
	PageTimeout = time.Second * 5

	// ServerShutdownTimeout is the timeout when doing a graceful server shutdown
	ServerShutdownTimeout = time.Second * 5

	// WebsocketPingPeriod is the interval between pings. Must be less than WebsocketPongTimeout.
	WebsocketPingPeriod = time.Second * 10

	// WebsocketPongTimeout is the time to wait for a pong from the client before cancelling
	WebsocketPongTimeout = time.Second * 20

	// WebsocketWriteTimeout is the write timeout for websockets
	WebsocketWriteTimeout = time.Second * 20

	// WebsocketInstructionTimeout is the time to wait for instructions from the client (e.g. during
	// playground compile)
	WebsocketInstructionTimeout = time.Second * 5

	// HttpTimeout is the time to wait for HTTP operations (e.g. getting meta data - not git)
	HttpTimeout = time.Second * 5

	ConcurrentStorageUploads = 10
)

var ValidExtensions = []string{".go", ".jsgo.html", ".inc.js", ".md"}

var Buckets = []string{Bucket[Src], Bucket[Pkg], Bucket[Index], Bucket[Git]}

var Static = []string{Src, Pkg, Index}
