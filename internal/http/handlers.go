package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/alee792/wonder/pkg/wonder"
)

// Index is the home page.
func (s *Server) Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello there")
	}
}

// Time returns the current time.
func (s *Server) Time() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", time.Now().Format(time.RFC1123Z))
	}
}

func (s *Server) Repo() http.HandlerFunc {
	type RepoRequest struct {
		Owner string
		Repo  string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req RepoRequest
		raw, err := ioutil.ReadAll(r.Body)
		if err != nil || raw == nil {
			http.Error(w, "valid POST body required", http.StatusBadRequest)
		}
		body := bytes.NewBuffer(raw)
		err = json.NewDecoder(body).Decode(&req)
		if err != nil {
			http.Error(w, fmt.Sprintf("json decode failed: %s", err), http.StatusBadRequest)
		}
		ctx := r.Context()
		repo, err := s.Wonder.ProcessRepo(ctx, req.Owner, req.Repo)
		if err != nil {
			http.Error(w, fmt.Sprintf("GetCommits failed: %s", err), http.StatusInternalServerError)
		}
		var uu []wonder.User
		for _, v := range repo.Users {
			uu = append(uu, *v)
		}
		err = json.NewEncoder(w).Encode(&uu)
		if err != nil {
			http.Error(w, fmt.Sprintf("json encode failed: %s", err), http.StatusInternalServerError)
		}
	}
}
