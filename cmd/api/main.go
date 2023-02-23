package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type key int

const keyServerAddr key = iota

type rootHandler struct{}

func (rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("static" + r.URL.Path)

	if err != nil {
		log.Println(err)
		return
	}

	if strings.HasSuffix(r.URL.Path, ".css") {
		w.Header().Add("Content-Type", "text/css; charset=utf-8")
	}

	io.Copy(w, f)
}

func initServer() {

	root_handler := rootHandler{}

	mux := http.NewServeMux()
	mux.Handle("/", root_handler)

	ctx, cancelCtx := context.WithCancel(context.Background())
	server := http.Server{
		Addr:         ":9630",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	log.Printf("Starting KVStoK Server on %s. Press CTRL-C to exit.", server.Addr)

	go func() {
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server is closed\n")
		} else if err != nil {
			fmt.Printf("Error starting server: %s\n", err.Error())
			os.Exit(1)
		}
		cancelCtx()
	}()
	<-ctx.Done()
}

func main() {
	initServer()
}
