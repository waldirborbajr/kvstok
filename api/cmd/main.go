package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nutsdb/nutsdb"
	"github.com/waldirborbajr/kvstok/api/internal"
	"github.com/waldirborbajr/kvstok/internal/database"
)

var store *database.Store

func main() {
	port := flag.Int("port", 8080, "HTTP port for the API server")
	masterPassword := flag.String("master", "", "Master password for kvstok")
	flag.Parse()

	if *masterPassword == "" {
		*masterPassword = os.Getenv("KVSTOK_MASTER_PASSWORD")
	}

	if *masterPassword == "" {
		log.Fatal("missing master password: set --master or KVSTOK_MASTER_PASSWORD")
	}

	var err error
	store, err = database.Init("")
	if err != nil {
		log.Fatalf("failed to open store: %v", err)
	}
	defer database.Close()

	if err := store.LoadMasterSalt(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("failed to load master salt: %v", err)
	}

	if err := store.SetMasterPassword(*masterPassword); err != nil {
		log.Fatalf("failed to set master password: %v", err)
	}

	if err := database.DB.Update(func(tx *nutsdb.Tx) error {
		return tx.NewBucket(nutsdb.DataStructureBTree, database.Bucket)
	}); err != nil && !strings.Contains(err.Error(), "already exists") {
		log.Fatalf("failed to initialize database bucket: %v", err)
	}

	srv := internal.NewServer(store)
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("🚀 Web server running at http://localhost:%d", *port)
	if err := srv.Start(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
