package http

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/middleware"

	"github.com/go-chi/chi"
)

// Routes sets a servers handlers.
func (s *Server) Routes() {
	s.Router.Route("/api", func(r chi.Router) {
		r.Get("/index.go", s.Index())
		r.Get("/time.go", s.Time())
		r.With(middleware.Logger).Post("/repo.go", s.Repo())
	})
	wd, _ := os.Getwd()
	filesDir := filepath.Join(wd, "web")
	FileServer(s.Router, "/", http.Dir(filesDir))
}
