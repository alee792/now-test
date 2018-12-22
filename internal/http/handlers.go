package http

import (
	"fmt"
	"net/http"
	"time"
)

// Index is the home page.
func (s *Server) Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello there")
	}
}

// Time returns the current time.
func (s *Server) Time() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", time.Now().Format(time.RFC1123Z))
	}
}
