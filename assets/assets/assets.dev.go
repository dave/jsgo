// +build dev

package assets

import (
	"net/http"

	"github.com/dave/patsy"
	"github.com/dave/patsy/vos"
)

// Assets contains project assets.
var Assets = func() http.FileSystem {
	dir, err := patsy.Dir(vos.Os(), "github.com/dave/jsgo/assets/assets/static")
	if err != nil {
		panic(err.Error())
	}
	return http.Dir(dir)
}()
