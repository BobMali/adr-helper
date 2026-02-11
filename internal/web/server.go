package web

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/go-chi/chi/v5"
)

// ServerOption configures optional Server behaviour.
type ServerOption func(*Server)

// WithFrontend enables serving an embedded SPA frontend.
// The provided fs.FS should contain index.html at its root.
func WithFrontend(frontend fs.FS) ServerOption {
	return func(s *Server) {
		s.frontend = frontend
	}
}

// Server holds the web server's dependencies and router.
type Server struct {
	router   chi.Router
	repo     adr.Repository
	frontend fs.FS
}

// NewServer creates a new Server with routes configured.
func NewServer(repo adr.Repository, opts ...ServerOption) *Server {
	r := chi.NewRouter()
	s := &Server{router: r, repo: repo}

	for _, opt := range opts {
		opt(s)
	}

	r.Get("/health", s.handleHealth)
	r.Get("/api/adr", s.handleListADRs)

	if s.frontend != nil {
		r.NotFound(spaHandler(s.frontend))
	}

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

type adrResponse struct {
	Number int        `json:"number"`
	Title  string     `json:"title"`
	Status adr.Status `json:"status"`
	Date   string     `json:"date"`
}

func (s *Server) handleListADRs(w http.ResponseWriter, r *http.Request) {
	if s.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}

	adrs, err := s.repo.List(r.Context())
	if err != nil {
		http.Error(w, "failed to list ADRs", http.StatusInternalServerError)
		return
	}

	resp := make([]adrResponse, len(adrs))
	for i, a := range adrs {
		dateStr := ""
		if !a.Date.IsZero() {
			dateStr = a.Date.Format("2006-01-02")
		}
		resp[i] = adrResponse{
			Number: a.Number,
			Title:  a.Title,
			Status: a.Status,
			Date:   dateStr,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// spaHandler returns an http.HandlerFunc that serves static files from the
// given fs.FS and falls back to index.html for unknown paths (SPA routing).
func spaHandler(frontend fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Clean the URL path and strip the leading slash.
		urlPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if urlPath == "" {
			urlPath = "index.html"
		}

		// Try opening the requested file.
		f, err := frontend.Open(urlPath)
		if err != nil {
			// File not found — serve index.html for SPA client-side routing.
			serveIndexHTML(w, frontend)
			return
		}
		f.Close()

		// Hashed assets get long-lived cache; everything else gets no-cache.
		if strings.HasPrefix(urlPath, "assets/") {
			w.Header().Set("Cache-Control", "public, immutable, max-age=31536000")
		}

		http.FileServer(http.FS(frontend)).ServeHTTP(w, r)
	}
}

func serveIndexHTML(w http.ResponseWriter, frontend fs.FS) {
	f, err := frontend.Open("index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	io.Copy(w, f.(io.Reader))
}
