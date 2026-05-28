package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// startWebServer starts the HTTP server with net/http.
func startWebServer(port int) {
	mux := http.NewServeMux()

	// ===== Middleware =====
	apiHandler := jwtMiddleware(http.HandlerFunc(apiRouter))

	// ===== API routes =====
	mux.Handle("/api/keys", apiHandler)
	mux.Handle("/api/keys/", apiHandler)

	// ===== Static assets =====
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/", http.StripPrefix("/", fs))

	// ===== Server =====
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

// apiRouter gerencia as rotas da API (substitui as rotas do Fiber)
func apiRouter(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/keys":
		handleListKeys(w, r)

	case r.Method == http.MethodPost && r.URL.Path == "/api/keys":
		handleAddKey(w, r)

	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/keys/"):
		handleDeleteKey(w, r)

	case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/keys/"):
		handleUpdateKey(w, r)

	default:
		http.NotFound(w, r)
	}
}

// jwtMiddleware validates a bearer token on incoming API requests.
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
