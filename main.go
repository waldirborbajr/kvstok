package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/waldirborbajr/kvstok/cmd"
	"github.com/waldirborbajr/kvstok/internal/database"
)

func main() {
	// Catch termination signals to close the database cleanly
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		if err := database.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to close database: %v\n", err)
		}
		os.Exit(0)
	}()

	cmd.Execute()

	// Ensure DB is closed on normal exit
	if err := database.Close(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to close database: %v\n", err)
	}
}
