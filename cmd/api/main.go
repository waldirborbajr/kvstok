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
	"time"

	"github.com/waldirborbajr/kvstok/internal/version"
)

type key int

const keyServerAddr key = iota

type rootHandler struct{}
type getkvHandler struct{}

func (rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("%s: got / request\n", ctx.Value(keyServerAddr))
	io.WriteString(w, "KVStoK (c) v"+version.AppVersion()+"\n")
}

func (getkvHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("%s: got /getkv request\n", ctx.Value(keyServerAddr))
	io.WriteString(w, "Hello, HTTP!\n")
}

func initServer() {

	root_handler := rootHandler{}
	getkv_handler := getkvHandler{}

	mux := http.NewServeMux()
	mux.Handle("/", root_handler)
	mux.Handle("/getkv", getkv_handler)

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
