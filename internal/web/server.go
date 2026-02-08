package web

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Server holds the web server's dependencies and router.
type Server struct {
	router chi.Router
}

// NewServer creates a new Server with routes configured.
func NewServer() *Server {
	r := chi.NewRouter()
	s := &Server{router: r}

	r.Get("/health", s.handleHealth)

	return s
}

// Handler returns the http.Handler for the server.
func (s *Server) Handler() http.Handler {
	return s.router
}

// ListenAndServe starts the HTTP server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
