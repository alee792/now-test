package main

import (
	"github.com/alee792/wonder/internal/http"
)

func main() {
	s := http.NewServer()
	s.Routes()
	s.ListenAndServe(8080, nil)
}
