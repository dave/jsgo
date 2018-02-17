package config

import "time"

const (

	// MaxConcurrentCompiles is the maximum number of concurrent compile jobs per server
	MaxConcurrentCompiles = 3

	// MaxQueue is the maximum queue length waiting for compile. After this an error is returned.
	MaxQueue = 100

	// ProjectId is the ID of the GCS project
	ProjectID = "jsgo-192815"

	CdnBucket   = "cdn.jsgo.io"
	IndexBucket = "jsgo.io"

	CdnHost   = "cdn.jsgo.io"
	IndexHost = "jsgo.io"

	CompileHost = "compile.jsgo.io"

	PkgDir = "pkg10"
	StdDir = "std10"

	AssetsFilename = "assets10.zip"

	// WriteTimeout is the timeout when serving static files
	WriteTimeout = time.Second * 2

	// CompileTimeout is the timeout when compiling a package.
	CompileTimeout = time.Second * 300

	// PageTimeout is the timeout when generating the compile page
	PageTimeout = time.Second * 5

	// ServerShutdownTimeout is the timeout when doing a graceful server shutdown
	ServerShutdownTimeout = time.Second * 5

	// WebsocketPingPeriod is the interval between pings
	WebsocketPingPeriod = time.Second

	// WebsocketPongTimeout is the time to wait for a pong from the client before cancelling
	WebsocketPongTimeout = time.Second * 2

	// WebsocketWriteTimeout is the write timeout for websockets
	WebsocketWriteTimeout = time.Second * 5
)
