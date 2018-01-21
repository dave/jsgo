package assets

import (
	"net/http"

	"github.com/dave/jsgo/assets/assets"
)

// Assets contains project assets.
var Assets http.FileSystem = assets.Assets
