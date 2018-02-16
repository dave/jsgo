package config

import "time"

// HashVersion is appended to the contents of all files when generating the SHA1 hash for the filename
const HashVersion byte = 1

// MaxConcurrentCompiles is the maximum number of concurrent compile jobs per server
const MaxConcurrentCompiles = 3

// MaxQueue is the maximum queue length waiting for compile. After this an error is returned.
const MaxQueue = 100

// ProjectId is the ID of the GCS project
const ProjectID = "jsgo-192815"

// WriteTimeout is the timeout when serving static files
const WriteTimeout = time.Second * 2

// CompileTimeout is the timeout when compiling a package.
const CompileTimeout = time.Second * 300

// PageTimeout is the timeout when generating the compile page
const PageTimeout = time.Second * 5

// ServerShutdownTimeout is the timeout when doing a graceful server shutdown
const ServerShutdownTimeout = time.Second * 5

// WebsocketPingPeriod is the interval between pings
const WebsocketPingPeriod = time.Second

// WebsocketPongTimeout is the time to wait for a pong from the client before cancelling
const WebsocketPongTimeout = time.Second * 2

// WebsocketWriteTimeout is the write timeout for websockets
const WebsocketWriteTimeout = time.Second * 5
