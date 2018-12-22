package handlers

import (
	"net/http"

	wonderhttp "github.com/alee792/wonder/internal/http"
)

var s *wonderhttp.Server

func init() {
	s = wonderhttp.NewServer()
}

// Index is a now compaitble wrapper.
func Index(w http.ResponseWriter, r *http.Request) {
	s.Index()(w, r)
}

// Time is a now compatible wrapper.
func Time(w http.ResponseWriter, r *http.Request) {
	s.Time()(w, r)
}
