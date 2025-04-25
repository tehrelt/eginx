package pool

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type reverseProxy struct {
	httputil.ReverseProxy
	forwardedBy string
	target      *url.URL
}

func newRP(target *url.URL) *reverseProxy {
	return &reverseProxy{
		forwardedBy:  "eginx",
		target:       target,
		ReverseProxy: *httputil.NewSingleHostReverseProxy(target),
	}
}

func (rp *reverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Forwarded-By", rp.forwardedBy)
	rp.ReverseProxy.ServeHTTP(w, r)
}
