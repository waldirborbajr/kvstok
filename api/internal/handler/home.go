// api/internal/handler/home.go
package handler

import (
	"net/http"
)

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":   "kvstok Web",
		"Version": "2025",
	}

	s.tmpl.ExecuteTemplate(w, "base.html", data)
}
