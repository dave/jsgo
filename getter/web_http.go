package getter

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/dave/jsgo/config"
	"golang.org/x/net/context/ctxhttp"
)

// httpClient is the default HTTP client, but a variable so it can be
// changed by tests, without modifying http.DefaultClient.
var httpClient = &http.Client{
	Timeout: config.HttpTimeout,
}

// impatientInsecureHTTPClient is used in -insecure mode,
// when we're connecting to https servers that might not be there
// or might be using self-signed certificates.
var impatientInsecureHTTPClient = &http.Client{
	Timeout: config.HttpTimeout,
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

type HTTPError struct {
	status     string
	StatusCode int
	url        string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s: %s", e.url, e.status)
}

// Get returns the data from an HTTP GET request for the given URL.
func Get(ctx context.Context, url string) ([]byte, error) {
	resp, err := ctxhttp.Get(ctx, httpClient, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err := &HTTPError{status: resp.Status, StatusCode: resp.StatusCode, url: url}

		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", url, err)
	}
	return b, nil
}

// GetMaybeInsecure returns the body of either the importPath's
// https resource or, if unavailable and permitted by the security mode, the http resource.
func GetMaybeInsecure(ctx context.Context, importPath string, insecure bool) (urlStr string, body io.ReadCloser, err error) {
	fetch := func(scheme string) (urlStr string, res *http.Response, err error) {
		u, err := url.Parse(scheme + "://" + importPath)
		if err != nil {
			return "", nil, err
		}
		u.RawQuery = "go-get=1"
		urlStr = u.String()
		if insecure && scheme == "https" { // fail earlier
			res, err = ctxhttp.Get(ctx, impatientInsecureHTTPClient, urlStr)
		} else {
			res, err = ctxhttp.Get(ctx, httpClient, urlStr)
		}
		return
	}
	closeBody := func(res *http.Response) {
		if res != nil {
			res.Body.Close()
		}
	}
	urlStr, res, err := fetch("https")
	if err != nil {
		if insecure {
			closeBody(res)
			urlStr, res, err = fetch("http")
		}
	}
	if err != nil {
		closeBody(res)
		return "", nil, err
	}
	// Note: accepting a non-200 OK here, so people can serve a
	// meta import in their http 404 page.
	return urlStr, res.Body, nil
}
