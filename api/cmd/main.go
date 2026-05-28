package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/nutsdb/nutsdb"
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
	store, err = database.NewStore("")
	if err != nil {
		log.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

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

	startWebServer(*port)
}

func startWebServer(port int) {
	mux := http.NewServeMux()

	apiHandler := jwtMiddleware(http.HandlerFunc(apiRouter))
	mux.Handle("/api/keys", apiHandler)
	mux.Handle("/api/keys/", apiHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"message": "kvstok API server"})
	})

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Server running at http://localhost:%d", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func apiRouter(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/keys":
		handleListKeys(w, r)

	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/keys/"):
		handleGetKey(w, r)

	case r.Method == http.MethodPost && r.URL.Path == "/api/keys":
		handleAddKey(w, r)

	case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/keys/"):
		handleUpdateKey(w, r)

	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/keys/"):
		handleDeleteKey(w, r)

	default:
		http.NotFound(w, r)
	}
}

type keyPayload struct {
	Key   string   `json:"key"`
	Value string   `json:"value"`
	TTL   uint32   `json:"ttl"`
	Tags  []string `json:"tags"`
}

func handleListKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := store.ListAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"keys": keys})
}

func handleGetKey(w http.ResponseWriter, r *http.Request) {
	key, err := keyFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	value, err := store.Get(key)
	if err != nil {
		handleStoreError(w, err)
		return
	}

	writeJSON(w, map[string]any{"key": key, "value": value})
}

func handleAddKey(w http.ResponseWriter, r *http.Request) {
	var payload keyPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	if payload.Key == "" || payload.Value == "" {
		http.Error(w, "key and value are required", http.StatusBadRequest)
		return
	}

	if err := store.Put(payload.Key, payload.Value, payload.TTL, payload.Tags); err != nil {
		handleStoreError(w, err)
		return
	}

	writeJSON(w, map[string]any{"status": "created", "key": payload.Key})
}

func handleUpdateKey(w http.ResponseWriter, r *http.Request) {
	key, err := keyFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payload keyPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	if payload.Value == "" {
		http.Error(w, "value is required", http.StatusBadRequest)
		return
	}

	if payload.Key != "" && payload.Key != key {
		http.Error(w, "payload key must match URL key", http.StatusBadRequest)
		return
	}

	if err := store.Put(key, payload.Value, payload.TTL, payload.Tags); err != nil {
		handleStoreError(w, err)
		return
	}

	writeJSON(w, map[string]any{"status": "updated", "key": key})
}

func handleDeleteKey(w http.ResponseWriter, r *http.Request) {
	key, err := keyFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := store.Delete(key); err != nil {
		handleStoreError(w, err)
		return
	}

	writeJSON(w, map[string]any{"status": "deleted", "key": key})
}

func keyFromPath(pathname string) (string, error) {
	const prefix = "/api/keys/"
	if !strings.HasPrefix(pathname, prefix) {
		return "", fmt.Errorf("invalid key path")
	}

	rawKey := strings.TrimPrefix(pathname, prefix)
	if rawKey == "" {
		return "", fmt.Errorf("key is required")
	}

	key, err := url.PathUnescape(rawKey)
	if err != nil {
		return "", fmt.Errorf("invalid key path: %w", err)
	}

	return key, nil
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func handleStoreError(w http.ResponseWriter, err error) {
	switch err {
	case database.ErrKeyNotFound:
		http.Error(w, "key not found", http.StatusNotFound)
	case database.ErrKeyExpired:
		http.Error(w, "key expired", http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if !validateBearerToken(token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validateBearerToken(token string) bool {
	const expectedToken = "kvstok-api-token"
	return token == expectedToken
}
