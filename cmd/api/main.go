package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/waldirborbajr/kvstok/internal/server"
)

func main() {

	server := &http.Server{
		Addr: ":9630",
		Handler: &server.WebApp{
			Routes: []server.Route{
				server.Route{
					Handler: server.HandlerIndex,
				},
			},
		},
	}

	server = nil

	go func() {
		if server != nil {
			log.Println("Starting KVStoK server")
			err := server.ListenAndServe()
			log.Println("HTTP server stopped: ", err)
		} else {
			log.Println("HTTP server temporarily disabled")
		}
	}()

	log.Println("KVStoK Server is now running. Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if server != nil {

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Println("Could not shutdown HTTP KVStoK server: ", err)
		}
	}

}
