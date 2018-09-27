package main

//go:generate go run -tags "dev local" ./initialise.go

// -tags ""            # PRODUCTION (Production Google Cloud endpoints)
// -tags "dev"         # DEVELOPMENT (Testing Google Cloud endpoints)
// -tags "dev local"   # LOCAL (Local mock endpoints)

// Add "-frizz" after "./initialise.go" for experimental Frizz additions
