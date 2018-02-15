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
// TODO: Why is this not used on the compile index page?
const WriteTimeout = time.Second * 2

// CompileTimeout is the timeout when compiling a package.
const CompileTimeout = time.Second * 300
