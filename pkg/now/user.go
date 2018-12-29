package main

import (
	"net/http"

	"github.com/alee792/wonder/pkg/now/handlers"
)

func User(w http.ResponseWriter, r *http.Request) {
	handlers.User(w, r)
}
