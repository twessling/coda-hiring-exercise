package pool

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type HandlerConfig struct {
	ListenAddr string
}

type PoolHandler struct {
	registerListenAddr string
	mux                *http.ServeMux
	clientPool         *ForwarderPool
}

func NewHandler(cfg *HandlerConfig, cp *ForwarderPool) *PoolHandler {
	ph := &PoolHandler{
		registerListenAddr: cfg.ListenAddr,
		mux:                http.NewServeMux(),
		clientPool:         cp,
	}

	ph.mux.HandleFunc(fmt.Sprintf("%s /", http.MethodPost), ph.registerClient)
	ph.mux.HandleFunc(fmt.Sprintf("%s /", http.MethodDelete), ph.deRegisterClient)
	return ph
}

func (ph *PoolHandler) ListenForClients(ctx context.Context) error {
	server := &http.Server{Addr: ph.registerListenAddr, Handler: ph.mux}

	// listen for context to stop server gracefully
	go func() {
		<-ctx.Done()
		log.Printf("Gracefully sutting down client listener...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Clietn listener shutdown Failed: %v", err)
		}
	}()

	return server.ListenAndServe()
}

func (ph *PoolHandler) registerClient(w http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// TODO: internal call, but what about validation? Just host/port valdiation? Full URI validation?

	addr := string(bytes)
	ph.clientPool.registerClient(addr)
}

func (ph *PoolHandler) deRegisterClient(w http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// TODO: internal call, but what about validation? Just host/port valdiation? Full URI validation?

	addr := string(bytes)
	ph.clientPool.deRegisterClient(addr)
}
