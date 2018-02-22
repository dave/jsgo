package main

import (
	"log"
	"net/http"

	githttp "github.com/AaronO/go-git-http"
	"github.com/dave/patsy"
	"github.com/dave/patsy/vos"
)

func main() {

	dir, err := patsy.Dir(vos.Os(), "github.com/dave/jsgo/testing/gitserver/roots")
	if err != nil {
		log.Fatal(err)
	}

	// Get git handler to serve a directory of repos
	git := githttp.New(dir)

	// Attach handler to http server
	http.Handle("/", git)

	// Start HTTP server
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
