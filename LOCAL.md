# Running locally

`compile.jsgo.io` and `play.jsgo.io` have a fully featured offline mode that can be used for offline 
development.

Offline mode simulates the Google Data Store and Google Storage Buckets by using a temporary 
directory, which defaults to `~/.jsgo-local`. You can change the location in [constants.go](https://github.com/dave/jsgo/blob/master/config/constants.go).

Instead of getting git repos from the internet, it uses the repos in your `GOPATH`, so any repo requested 
that's not in your `GOPATH` will fail. 

| Production | Local |
| --- | --- |
| git from the internet | repos in your GOPATH |
| google datastore | json files in temporary dir |
| google storage | files in temporary dir |

### Setup

Get the latest source:

`go get -u github.com/dave/jsgo/...`

Initialise the project:

`cd $GOPATH/src/github.com/dave/jsgo/initialise`

`go generate`

Start the server: 

`cd $GOPATH/src/github.com/dave/jsgo/server/main`

`go run -tags "norwfs dev local" main.go`

Open a browser and head to [localhost:8080](http://localhost:8080/) to open the jsgo playground.

This will also start some other servers:

| Local | Production equivalent |
| --- | --- |
| localhost:8080 | play.jsgo.io |
| localhost:8081 | compile.jsgo.io |
| localhost:8082 | frizz.io |
| localhost:8083 | wasmgo.jsgo.io |
| localhost:8091 | src.jsgo.io |
| localhost:8092 | pkg.jsgo.io |
| localhost:8093 | jsgo.io |

