package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alee792/wonder/pkg/wonder"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// Config for Server.
type Config struct {
	Wonder wonder.Config
}

// Server wraps the default net/http implementation.
type Server struct {
	Wonder *wonder.Server
	Router chi.Router
	Logger *zap.SugaredLogger
	http   http.Server
}

// NewServer for HTTP.
func NewServer(cfg Config) *Server {
	r := chi.NewRouter()
	return &Server{
		Wonder: wonder.NewClient(cfg.Wonder),
		Router: r,
		Logger: zap.NewExample().Sugar(),
		http: http.Server{
			Handler: r,
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

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
