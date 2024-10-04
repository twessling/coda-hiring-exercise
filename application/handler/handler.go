package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const handledByHeader = "X-Handled-By"

type Config struct {
	Addr string
	Id   string
}

type Handler struct {
	addr string
	id   string
	mux  *http.ServeMux
}

func New(cfg *Config) *Handler {
	h := &Handler{
		addr: cfg.Addr,
		mux:  http.NewServeMux(),
		id:   cfg.Id,
	}

	h.mux.HandleFunc(fmt.Sprintf("%s /json", http.MethodPost), h.handlePostJson)

	return h
}

func (h *Handler) ListenAndServe(ctx context.Context) error {
	server := &http.Server{Addr: h.addr, Handler: h.mux}

	// listen for context to stop server gracefully
	go func() {
		<-ctx.Done()
		log.Printf("Gracefully shutting down handler..")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Handler shutdown failed: %v", err)
		}
	}()

	return server.ListenAndServe()
}

func (h *Handler) handlePostJson(w http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !json.Valid(bytes) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add(handledByHeader, h.id)
	_, err = w.Write(bytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
