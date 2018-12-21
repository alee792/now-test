package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// Server wraps the default net/http implementation.
type Server struct {
	Router chi.Router
	Logger *zap.SugaredLogger
	http   http.Server
}

// NewServer for HTTP.
func NewServer() *Server {
	return &Server{
		Logger: zap.NewExample().Sugar(),
		http: http.Server{
			Handler: chi.NewRouter(),
		},
	}
}

// ListenAndServe on a port using a specified handler.
func (s *Server) ListenAndServe(port int, r http.Handler) error {
	if r == nil {
		r = s.Router
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
