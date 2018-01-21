package main

import (
	"log"
	"net/http"

	"path/filepath"

	"os"

	"github.com/dave/patsy"
	"github.com/dave/patsy/vos"
	"github.com/shurcooL/vfsgen"
)

func main() {

	dir, err := patsy.Dir(vos.Os(), "github.com/dave/jsgo/assets/assets")
	if err != nil {
		log.Fatalln(err)
	}

	staticDir := filepath.Join(dir, "static")
	assetFilepath := filepath.Join(dir, "assets.go")
	tempFilepath := filepath.Join(dir, "assets.go.tmp")

	options := vfsgen.Options{
		PackageName:  "assets",
		BuildTags:    "!dev",
		VariableName: "Assets",
		Filename:     tempFilepath,
	}
	if err := vfsgen.Generate(http.Dir(staticDir), options); err != nil {
		log.Fatalln(err)
	}
	if err := os.Remove(assetFilepath); err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
	}
	if err := os.Rename(tempFilepath, assetFilepath); err != nil {
		log.Fatalln(err)
	}
}
