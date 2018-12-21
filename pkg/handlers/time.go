package handlers

import (
	"fmt"
	"net/http"
	"time"
)

func Time(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", time.Now().Format(time.RFC1123Z))
}
