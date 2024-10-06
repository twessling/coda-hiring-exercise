package pool

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type stage int

const (
	stage_ok stage = iota
	stage_slow
	stage_dead
)

type Forwarder interface {
	Forward(w http.ResponseWriter, req *http.Request)
	Host() string
}
type forwardHandler struct {
	addr         string
	proxy        *httputil.ReverseProxy
	movingWindow *movingWindow
	weight       float64
}

func newForwardHandler(addr string) Forwarder {
	uri, _ := url.Parse(fmt.Sprintf("http://%s", addr)) // TODO: should the 'http://' be here or in the client's registration data?
	proxy := httputil.NewSingleHostReverseProxy(uri)
	return &forwardHandler{
		addr:         addr,
		proxy:        proxy,
		movingWindow: newMovingWindow(),
		weight:       1,
	}
}

func (h *forwardHandler) Forward(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	h.proxy.ServeHTTP(w, req)
	duration := time.Since(start)

	h.movingWindow.trackNewDuration(duration)
}

func (h *forwardHandler) Host() string {
	return h.addr
}
