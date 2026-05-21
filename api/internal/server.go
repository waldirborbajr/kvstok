package internal

import (
	"html/template"
	"net/http"

	"github.com/waldirborbajr/kvstok/api/internal/template"
	"github.com/waldirborbajr/kvstok/internal/database"
)

type Server struct {
	mux   *http.ServeMux
	tmpl  *template.Template
	store *database.Store // seu store com segurança
}

func NewServer(store *database.Store) *Server {
	s := &Server{
		mux:   http.NewServeMux(),
		store: store,
	}

	// Carrega todos os templates
	tmpl, err := template.ParseFS(template.Files, "html/*.html", "html/**/*.html")
	if err != nil {
		panic(err)
	}
	s.tmpl = tmpl

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Página inicial
	s.mux.HandleFunc("/", s.handleHome)

	// API endpoints com HTMX
	s.mux.HandleFunc("/kv/", s.handleKV)
	s.mux.HandleFunc("/search", s.handleSearch)

	// Static files (CSS, HTMX, etc.)
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(template.StaticFiles))))
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}
