package router

import (
	"context"
	"fmt"
	"log"
	"mrbarrel/router/pool"
	"net/http"
	"time"
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

func New(cfg *Config, clientPool *pool.ClientPool) *Router {
	r := &Router{
		addr:    cfg.Addr,
		clients: clientPool,
		mux:     http.NewServeMux(),
	}

	r.mux.HandleFunc(fmt.Sprintf("%s /", http.MethodPost), r.handle)

	return r
}

func (r *Router) ListenAndServe(ctx context.Context) error {
	server := &http.Server{Addr: r.addr, Handler: r.mux}

	// listen for context to stop server gracefully
	go func() {
		<-ctx.Done()
		log.Printf("INFO: Gracefully sutting down router...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Router shutdown failed: %v", err)
		}
	}()

	return server.ListenAndServe()
}

func (r *Router) handle(w http.ResponseWriter, req *http.Request) {
	forwarder, err := r.clients.Next()
	if err != nil {
		log.Printf("could not get client: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	forwarder.Forward(w, req)
}
