package main

import (
	"fmt"

	"github.com/alee792/wonder/internal/http"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	var cfg http.Config
	envconfig.MustProcess("", &cfg)
	fmt.Printf("%+v", cfg)
	s := http.NewServer(cfg)
	s.Routes()
	s.ListenAndServe(8080, nil)
}
