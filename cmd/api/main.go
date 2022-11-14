package main

import (
	"context"
	"fmt"
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

	server = nil

	if server != nil {
		log.Println("Starting KVStoK server")
		err := server.ListenAndServe()
		log.Println("HTTP server stopped: ", err)
	} else {
		log.Println("HTTP server temporarily disabled")
	}

	log.Println("KVStoK Server is now running. Press CTRL-C to exit.")

	if server != nil {

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Println("Could not shutdown HTTP KVStoK server: ", err)
		}
	}

}

func main() {

	fmt.Println("Starting KVStoK server")

	ch := make(chan struct{})
	go initServer()
	<-ch

}
