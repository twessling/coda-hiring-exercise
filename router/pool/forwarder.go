package pool

import (
	"fmt"
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
	rateLimiter *rateLimiter
	weight      float64
}

func newForwardHandler(addr string) Forwarder {
	uri, _ := url.Parse(fmt.Sprintf("http://%s", addr)) // TODO: should the 'http://' be here or in the client's registration data?
	proxy := httputil.NewSingleHostReverseProxy(uri)
	return &forwardHandler{
		addr:        addr,
		proxy:       proxy,
		rateLimiter: newRateLimiter(),
		weight:      1,
	}
}

func (h *forwardHandler) Forward(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	h.proxy.ServeHTTP(w, req)
	duration := time.Since(start)

	h.rateLimiter.trackNewDuration(duration)
}

func (h *forwardHandler) Host() string {
	return h.addr
}

func (h *forwardHandler) CanForward() bool {
	return h.rateLimiter.canHandleCall()
}
