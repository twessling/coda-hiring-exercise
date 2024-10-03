package router

import (
	"fmt"
	"log"
	"mrbarrel/router/pool"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Config struct {
	Addr       string
	clientPool *pool.ClientPool
}

type Router struct {
	addr    string
	clients *pool.ClientPool
	mux     *http.ServeMux
}

func New(cfg *Config) *Router {
	r := &Router{
		addr:    cfg.Addr,
		clients: cfg.clientPool,
		mux:     http.NewServeMux(),
	}

	r.mux.HandleFunc(fmt.Sprintf("%s /", http.MethodPost), r.handle)

	return r
}

func (r *Router) ListenAndServe() error {
	return http.ListenAndServe(r.addr, r.mux)
}

func (r *Router) handle(w http.ResponseWriter, req *http.Request) {
	clientAddr, err := r.clients.Next()
	if err != nil {
		log.Printf("could not get client: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	uri, _ := url.Parse(fmt.Sprintf("http://%s", clientAddr))
	proxy := httputil.NewSingleHostReverseProxy(uri)
	proxy.ServeHTTP(w, req)
}
