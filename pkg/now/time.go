package main

import (
	"net/http"

	"github.com/alee792/wonder/pkg/handlers"
)

func Time(w http.ResponseWriter, r *http.Request) {
	handlers.Time(w, r)
}
