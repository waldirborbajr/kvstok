package internal

import (
	"fmt"
	"html"
	htmltmpl "html/template"
	"net/http"
	"sort"
	"strings"

	"github.com/waldirborbajr/kvstok/api/internal/template"
	"github.com/waldirborbajr/kvstok/internal/database"
)

type Server struct {
	mux   *http.ServeMux
	tmpl  *htmltmpl.Template
	store *database.Store
}

func NewServer(store *database.Store) *Server {
	s := &Server{
		mux:   http.NewServeMux(),
		store: store,
	}

	// Load all HTML templates
	tmpl, err := htmltmpl.ParseFS(template.Files, "html/*.html")
	if err != nil {
		panic(err)
	}
	s.tmpl = tmpl

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Static assets served from embed
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(template.StaticFiles))))

	// Página inicial
	s.mux.HandleFunc("/", s.handleHome)

	// API endpoints com HTMX
	s.mux.HandleFunc("/api/keys", s.handleKeys)
	s.mux.HandleFunc("/api/search", s.handleSearch)
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) handleKeys(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListKeys(w, r)
	case http.MethodPost:
		s.handleAddKey(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		s.handleListKeys(w, r)
		return
	}

	entries, err := s.store.Search(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.renderKeyList(w, entries)
}

func (s *Server) handleListKeys(w http.ResponseWriter, r *http.Request) {
	entries, err := s.store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.renderKeyList(w, entries)
}

func (s *Server) handleAddKey(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form payload", http.StatusBadRequest)
		return
	}

	key := strings.TrimSpace(r.FormValue("key"))
	value := strings.TrimSpace(r.FormValue("value"))
	tags := parseTags(r.FormValue("tags"))

	if key == "" || value == "" {
		http.Error(w, "key and value are required", http.StatusBadRequest)
		return
	}

	if err := s.store.Put(key, value, 0, tags); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, entry, err := s.store.GetRaw(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.renderKeyCard(w, key, *entry)
}

func (s *Server) renderKeyList(w http.ResponseWriter, entries map[string]database.SecretEntry) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if len(entries) == 0 {
		fmt.Fprint(w, `<div class="notification is-warning">Nenhuma chave encontrada.</div>`)
		return
	}

	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		s.renderKeyCard(w, key, entries[key])
	}
}

func (s *Server) renderKeyCard(w http.ResponseWriter, key string, entry database.SecretEntry) {
	value := html.EscapeString(entry.Value)
	tags := html.EscapeString(strings.Join(entry.Tags, ", "))

	fmt.Fprintf(w, `<article class="box key-card">
		<div class="level">
			<div class="level-left">
				<div>
					<p class="title is-6">%s</p>
					<p class="subtitle is-7"><code class="value">%s</code></p>
				</div>
			</div>
		</div>
		`, html.EscapeString(key), value)

	if tags != "" {
		fmt.Fprintf(w, `<div class="tags"><span class="tag is-light">%s</span></div>`, tags)
	}

	fmt.Fprintf(w, `
	</article>`)
}

func parseTags(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
