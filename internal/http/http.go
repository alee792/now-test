package http

import (
	"path/filepath"
	"os"
	"github.com/alee792/wonder/pkg/handlers"
	"fmt"
	"net/http"
	"strings"

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
	r := chi.NewRouter()
	return &Server{
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

// Routes sets a servers handlers.
func (s *Server) Routes() {
	s.Router.Route("/api", func(r chi.Router){
		r.Get("/index.go", func(w http.ResponseWriter, r *http.Request){ 
			handlers.Index(w, r)
		})
		r.Get("/time.go", func(w http.ResponseWriter, r *http.Request){
			handlers.Time(w, r)
		})
	})
	wd, _ := os.Getwd()
	filesDir := filepath.Join(wd, "web")
	FileServer(s.Router, "/", http.Dir(filesDir))
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