package pool

import (
	"fmt"
	"mrbarrel/router/pool/ratelimit"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Forwarder interface {
	Forward(w http.ResponseWriter, req *http.Request)
	Host() string
	CanForward() bool
}
type forwardHandler struct {
	addr        string
	proxy       *httputil.ReverseProxy
	rateLimiter *ratelimit.RateLimiter
}

func newForwardHandler(addr string) Forwarder {
	uri, _ := url.Parse(fmt.Sprintf("http://%s", addr)) // TODO: should the 'http://' be here or in the client's registration data?
	proxy := httputil.NewSingleHostReverseProxy(uri)
	return &forwardHandler{
		addr:        addr,
		proxy:       proxy,
		rateLimiter: ratelimit.NewRateLimiter(),
	}
}

func (h *forwardHandler) Forward(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	h.proxy.ServeHTTP(w, req)
	duration := time.Since(start)

	h.rateLimiter.TrackNewDuration(duration)
}

func (h *forwardHandler) Host() string {
	return h.addr
}

func (h *forwardHandler) CanForward() bool {
	return h.rateLimiter.CanHandleCall()
}
