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
}
type poolEntry struct {
	Addr         string
	movingWindow *movingWindow
	weight       float64
}

func (h *poolEntry) Forward(w http.ResponseWriter, req *http.Request) {
	uri, _ := url.Parse(fmt.Sprintf("http://%s", h.Addr)) // TODO: should the 'http://' be here or in the client's registration data?
	proxy := httputil.NewSingleHostReverseProxy(uri)

	start := time.Now()
	proxy.ServeHTTP(w, req)
	duration := time.Since(start)

	h.movingWindow.trackNewDuration(duration)
}
