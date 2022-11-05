package server

import (
	"fmt"
	"net/http"
)

type Route struct {
	Handler func(wp *WebApp, w http.ResponseWriter, r *http.Request)
}

type WebApp struct {
	Routes []Route
}

func HandlerIndex(wp *WebApp, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "KVStoK - /")
}
