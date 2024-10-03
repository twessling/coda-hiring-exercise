package handler

import (
	"fmt"
	"io"
	"net/http"
)

type Config struct {
	Addr string
}

type Handler struct {
	addr string
	mux  *http.ServeMux
}

func New(cfg *Config) *Handler {
	h := &Handler{
		addr: cfg.Addr,
		mux:  http.NewServeMux(),
	}

	h.mux.HandleFunc(fmt.Sprintf("%s /json", http.MethodPost), h.handlePostJson)

	return h
}

func (h *Handler) ListenAndServe() error {
	return http.ListenAndServe(h.addr, h.mux)
}

func (j *Handler) handlePostJson(w http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = validateJson(bytes)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = w.Write(bytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
