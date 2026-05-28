package internal

import (
	"net/http"
)

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":   "kvstok Web",
		"Version": "2025",
	}

	err := s.tmpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
