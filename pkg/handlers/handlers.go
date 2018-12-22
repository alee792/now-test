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
	s.Index()(w, r)
}

// Time is a now compatible wrapper.
func Time(w http.ResponseWriter, r *http.Request) {
	s.Time()(w, r)
}

// User is a now compatible wrapper.
func User(w http.ResponseWriter, r *http.Request) {
	s.UsersByRepo()(w, r)
}
