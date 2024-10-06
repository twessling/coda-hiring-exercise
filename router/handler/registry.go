package router

import (
	"context"
	"fmt"
	"io"
	"log"
	"mrbarrel/router/pool"
	"net/http"
	"time"
)

type RegistryHandlerConfig struct {
	ListenAddr string
}

type RegistryHandler struct {
	registerListenAddr string
	mux                *http.ServeMux
	clientRegistrar    pool.ClientRegistrar
}

func NewRegistryHandler(cfg *RegistryHandlerConfig, cp pool.ClientRegistrar) *RegistryHandler {
	ph := &RegistryHandler{
		registerListenAddr: cfg.ListenAddr,
		mux:                http.NewServeMux(),
		clientRegistrar:    cp,
	}

	ph.mux.HandleFunc(fmt.Sprintf("%s /", http.MethodPost), ph.registerClient)
	ph.mux.HandleFunc(fmt.Sprintf("%s /", http.MethodDelete), ph.deRegisterClient)
	return ph
}

func (ph *RegistryHandler) ListenForClients(ctx context.Context) error {
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

func (ph *RegistryHandler) registerClient(w http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// TODO: internal call, but what about validation? Just host/port valdiation? Full URI validation?

	addr := string(bytes)
	ph.clientRegistrar.RegisterClient(addr)
}

func (ph *RegistryHandler) deRegisterClient(w http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// TODO: internal call, but what about validation? Just host/port valdiation? Full URI validation?

	addr := string(bytes)
	ph.clientRegistrar.DeRegisterClient(addr)
}
