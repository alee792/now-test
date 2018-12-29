package handlers

import (
	"net/http"
	"time"

	wonderhttp "github.com/alee792/wonder/internal/http"
	"github.com/kelseyhightower/envconfig"
)

var s *wonderhttp.Server

func init() {
	var cfg wonderhttp.Config
	envconfig.MustProcess("", &cfg)
	s = wonderhttp.NewServer(cfg)
	s.Logger.Info("new instance at %s", time.Now().Format(time.RFC1123Z))
}

// Index is a now compaitble wrapper.
func Index(w http.ResponseWriter, r *http.Request) {
	setCacheControlHeader(w)
	s.Index()(w, r)
}

// Time is a now compatible wrapper.
func Time(w http.ResponseWriter, r *http.Request) {
	setCacheControlHeader(w)
	s.Time()(w, r)
}

// Repo is a now compatible wrapper.
func Repo(w http.ResponseWriter, r *http.Request) {
	setCacheControlHeader(w)
	s.Repo()(w, r)
}

func setCacheControlHeader(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "s-max-age=86400, max-age=86400")
}
