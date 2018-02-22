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

	// WebsocketPingPeriod is the interval between pings. Must be less than WebsocketPongTimeout.
	WebsocketPingPeriod = time.Second * 10

	// WebsocketPongTimeout is the time to wait for a pong from the client before cancelling
	WebsocketPongTimeout = time.Second * 20

	// WebsocketWriteTimeout is the write timeout for websockets
	WebsocketWriteTimeout = time.Second * 20

	// GitCloneTimeout is the time to wait for a git clone operation
	GitCloneTimeout = time.Second * 60

	// GitPullTimeout is the time to wait for a git pull operation
	GitPullTimeout = time.Second * 60

	// GitListTimeout is the time to wait for a git list operation
	GitListTimeout = time.Second * 10

	// GitMaxBytes is the maximum bytes allowed to be used by the git clone operation before returning
	// an error.
	GitMaxBytes = 50 * 1024 * 1024 // 100 MB

	// GitMaxRefs is the maximum refs returned by git ls-remote before the repo is deemed too big (this check is currently disabled)
	GitMaxRefs = 10000

	// GitMaxObjects is the maximum objects in git clone progress
	GitMaxObjects = 10000

	// HttpTimeout is the time to wait for HTTP operations (e.g. getting meta data - not git)
	HttpTimeout = time.Second * 5
)
