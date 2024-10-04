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
	clientPool         *ClientPool
}

func NewHandler(cfg *HandlerConfig, cp *ClientPool) *PoolHandler {
	ph := &PoolHandler{
		registerListenAddr: cfg.ListenAddr,
		mux:                http.NewServeMux(),
		clientPool:         cp,
	}

	ph.mux.HandleFunc(fmt.Sprintf("%s /", http.MethodPost), ph.registerClient)
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
	addr := string(bytes)
	ph.clientPool.registerClient(addr)
}
