package main

import (
	"net/http"

	"github.com/alee792/wonder/pkg/now/handlers"
)

func Index(w http.ResponseWriter, r *http.Request) {
	handlers.Index(w, r)
}
