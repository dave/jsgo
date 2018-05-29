package server

import (
	"net/http"
	"strings"

	"fmt"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/frizz"
	"github.com/dave/jsgo/server/jsgo"
	"github.com/dave/jsgo/server/play"
)

type pageType int

const (
	UnknownPage pageType = iota
	PlayPage
	JsgoPage
	FrizzPage
)

func getPage(req *http.Request) pageType {
	if config.DEV {
		switch {
		case strings.HasSuffix(req.Host, "8080"):
			return PlayPage
		case strings.HasSuffix(req.Host, "8081"):
			return JsgoPage
		case strings.HasSuffix(req.Host, "8082"):
			return FrizzPage
		}
	} else {
		switch req.Host {
		case "play.jsgo.io":
			return PlayPage
		case "compile.jsgo.io":
			return JsgoPage
		case "frizz.io":
			return FrizzPage
		}
	}
	return UnknownPage
}

func (h *Handler) PageHandler(w http.ResponseWriter, req *http.Request) {
	switch getPage(req) {
	case PlayPage:
		play.Page(w, req, h.Database)
		return
	case JsgoPage:
		jsgo.Page(w, req, h.Database)
		return
	case FrizzPage:
		frizz.Page(w, req, h.Database)
		return
	default:
		http.Error(w, fmt.Sprintf("unknown host %s", req.Host), 500)
		return
	}
}
