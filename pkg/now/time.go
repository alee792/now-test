package main

import (
	"net/http"
)

func Time(w http.ResponseWriter, r *http.Request) {
	s.Time()(w, r)
}
