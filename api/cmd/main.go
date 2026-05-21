package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// startWebServer inicia o servidor HTTP usando apenas net/http
func startWebServer(port int) {
	mux := http.NewServeMux()

	// ====================== MIDDLEWARE ======================
	// Aplicar middleware de autenticação JWT em todas as rotas da API
	apiHandler := jwtMiddleware(http.HandlerFunc(apiRouter))

	// ====================== ROTAS API ======================
	mux.Handle("/api/keys", apiHandler)
	mux.Handle("/api/keys/", apiHandler) // para suportar /api/keys/:key

	// ====================== ARQUIVOS ESTÁTICOS ======================
	// Serve arquivos do diretório ./web/static
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/", http.StripPrefix("/", fs))

	// ====================== SERVIDOR ======================
	addr := fmt.Sprintf(":%d", port)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Servidor rodando em http://localhost:%d", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
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

// ====================== MIDDLEWARE JWT (exemplo) ======================
func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: implemente sua lógica de JWT aqui
		token := r.Header.Get("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// token = strings.TrimPrefix(token, "Bearer ")
		// validar token...

		next.ServeHTTP(w, r)
	})
}
