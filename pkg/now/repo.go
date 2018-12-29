package main

import (
	"net/http"

	"github.com/alee792/wonder/pkg/now/handlers"
)

func Repo(w http.ResponseWriter, r *http.Request) {
	handlers.Repo(w, r)
}
