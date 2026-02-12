package web

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/go-chi/chi/v5"
)

// StatusUpdater can change an ADR's status and return the updated record.
type StatusUpdater interface {
	UpdateStatus(ctx context.Context, number int, newStatus string) (*adr.ADR, error)
}

// Superseder performs bidirectional supersede updates between two ADRs.
type Superseder interface {
	Supersede(ctx context.Context, supersededNum, supersedingNum int) (*adr.ADR, error)
}

// ServerOption configures optional Server behaviour.
type ServerOption func(*Server)

// WithFrontend enables serving an embedded SPA frontend.
// The provided fs.FS should contain index.html at its root.
func WithFrontend(frontend fs.FS) ServerOption {
	return func(s *Server) {
		s.frontend = frontend
	}
}

// WithStatusUpdater enables the PATCH status endpoint.
func WithStatusUpdater(u StatusUpdater) ServerOption {
	return func(s *Server) {
		s.updater = u
	}
}

// WithSuperseder enables the supersede flow in the PATCH status endpoint.
func WithSuperseder(sup Superseder) ServerOption {
	return func(s *Server) {
		s.superseder = sup
	}
}

// Server holds the web server's dependencies and router.
type Server struct {
	router     chi.Router
	repo       adr.Repository
	frontend   fs.FS
	updater    StatusUpdater
	superseder Superseder
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
	r.Get("/api/adr/statuses", s.handleStatuses)
	r.Get("/api/adr/{number}", s.handleGetADR)
	r.Patch("/api/adr/{number}/status", s.handleUpdateStatus)

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

type adrDetailResponse struct {
	Number  int        `json:"number"`
	Title   string     `json:"title"`
	Status  adr.Status `json:"status"`
	Date    string     `json:"date"`
	Content string     `json:"content"`
}

func toResponse(a adr.ADR) adrResponse {
	dateStr := ""
	if !a.Date.IsZero() {
		dateStr = a.Date.Format("2006-01-02")
	}
	return adrResponse{
		Number: a.Number,
		Title:  a.Title,
		Status: a.Status,
		Date:   dateStr,
	}
}

func toDetailResponse(a adr.ADR) adrDetailResponse {
	dateStr := ""
	if !a.Date.IsZero() {
		dateStr = a.Date.Format("2006-01-02")
	}
	return adrDetailResponse{
		Number:  a.Number,
		Title:   a.Title,
		Status:  a.Status,
		Date:    dateStr,
		Content: a.Content,
	}
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

	if q := r.URL.Query().Get("q"); q != "" {
		adrs = adr.FilterByQuery(adrs, q)
	}

	resp := make([]adrResponse, len(adrs))
	for i, a := range adrs {
		resp[i] = toResponse(a)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleGetADR(w http.ResponseWriter, r *http.Request) {
	if s.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}

	numberStr := chi.URLParam(r, "number")
	number, err := strconv.Atoi(numberStr)
	if err != nil || number <= 0 {
		http.Error(w, "invalid ADR number", http.StatusBadRequest)
		return
	}

	record, err := s.repo.Get(r.Context(), number)
	if err != nil {
		if errors.Is(err, adr.ErrNotFound) {
			http.Error(w, "ADR not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get ADR", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(toDetailResponse(*record)); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleStatuses(w http.ResponseWriter, _ *http.Request) {
	statuses := adr.AllStatuses()
	names := make([]string, len(statuses))
	for i, st := range statuses {
		names[i] = st.String()
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(names); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if s.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}

	if s.updater == nil {
		http.Error(w, "status updates not supported", http.StatusNotImplemented)
		return
	}

	numberStr := chi.URLParam(r, "number")
	number, err := strconv.Atoi(numberStr)
	if err != nil || number <= 0 {
		http.Error(w, "invalid ADR number", http.StatusBadRequest)
		return
	}

	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var body struct {
		Status       string `json:"status"`
		SupersededBy *int   `json:"supersededBy,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	parsed, ok := adr.ParseStatus(body.Status)
	if !ok {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	var record *adr.ADR
	if parsed == adr.Superseded {
		if body.SupersededBy == nil {
			http.Error(w, "supersededBy is required when status is Superseded", http.StatusBadRequest)
			return
		}
		if s.superseder == nil {
			http.Error(w, "supersede not supported", http.StatusNotImplemented)
			return
		}
		record, err = s.superseder.Supersede(r.Context(), number, *body.SupersededBy)
	} else {
		record, err = s.updater.UpdateStatus(r.Context(), number, body.Status)
	}
	if err != nil {
		if errors.Is(err, adr.ErrNotFound) {
			http.Error(w, "ADR not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(toDetailResponse(*record)); err != nil {
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
