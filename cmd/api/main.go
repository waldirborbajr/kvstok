package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

type serverHandler struct{}

func initServer() {

	server := &http.Server{
		Addr:         ":9630",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Printf("Starting KVStoK Server on %s. Press CTRL-C to exit.", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("HTTP server stopped: %s", err.Error())
	}

	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Println("Could not shutdown HTTP KVStoK server: ", err)
		}
	}

}

func main() {
	ch := make(chan struct{})
	go initServer()
	<-ch

}
