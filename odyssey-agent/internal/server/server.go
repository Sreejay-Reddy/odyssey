package server

import (
	"fmt"
	"context"
	"net/http"
)

type Server struct {
	mux    *http.ServeMux
	server *http.Server
}

func New(addr string) *Server {
	mux := http.NewServeMux()

	s := &Server{
		mux: mux,
	}

	mux.HandleFunc("/health", s.health)

	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}