package main

import (
	"net/http"

	wonderhttp "github.com/alee792/wonder/pkg/http"
)

var s *wonderhttp.Server

func init() {
	s = wonderhttp.NewServer()
}

func main() {}

func Index(w http.ResponseWriter, r *http.Request) {
	s.Index()(w, r)
}
