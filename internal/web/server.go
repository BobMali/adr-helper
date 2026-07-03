package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
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

// ContentUpdater can replace the full markdown content of an ADR.
type ContentUpdater interface {
	UpdateContent(ctx context.Context, number int, content string) (*adr.ADR, error)
}

// Relator adds bidirectional relation links between two ADRs.
type Relator interface {
	AddRelation(ctx context.Context, sourceNum, targetNum int) (*adr.ADR, error)
}

// ScopeStore reads and extends the project's scope vocabulary, persisting
// additions. Implementations must be safe for concurrent use.
type ScopeStore interface {
	Scopes() []string
	AddScope(value string) ([]string, error)
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

// WithRelator enables the relation endpoint.
func WithRelator(rel Relator) ServerOption {
	return func(s *Server) {
		s.relator = rel
	}
}

// WithContentUpdater enables the PUT content endpoint.
func WithContentUpdater(u ContentUpdater) ServerOption {
	return func(s *Server) {
		s.contentUpdater = u
	}
}

// WithConfig provides the project configuration for template rendering.
func WithConfig(cfg *adr.Config) ServerOption {
	return func(s *Server) {
		s.config = cfg
	}
}

// WithScopeStore enables the scope vocabulary endpoints.
func WithScopeStore(store ScopeStore) ServerOption {
	return func(s *Server) {
		s.scopeStore = store
	}
}

// Saver can persist a new ADR record.
type Saver interface {
	Save(ctx context.Context, record *adr.ADR) error
	NextNumber(ctx context.Context) (int, error)
}

// Server holds the web server's dependencies and router.
type Server struct {
	router         chi.Router
	repo           adr.Repository
	frontend       fs.FS
	updater        StatusUpdater
	superseder     Superseder
	relator        Relator
	contentUpdater ContentUpdater
	scopeStore     ScopeStore
	config         *adr.Config
}

// NewServer creates a new Server with routes configured.
func NewServer(repo adr.Repository, opts ...ServerOption) *Server {
	r := chi.NewRouter()
	s := &Server{router: r, repo: repo}

	for _, opt := range opts {
		opt(s)
	}

	r.Get("/health", s.handleHealth)
	r.Get("/api/config", s.handleGetConfig)
	r.Get("/api/template-sections", s.handleGetTemplateSections)
	r.Get("/api/scopes", s.handleGetScopes)
	r.Post("/api/scopes", s.handleAddScope)
	r.Get("/api/adr", s.handleListADRs)
	r.Get("/api/adr/statuses", s.handleStatuses)
	r.Post("/api/adr", s.handleCreateADR)
	r.Get("/api/adr/{number}", s.handleGetADR)
	r.Put("/api/adr/{number}", s.handleUpdateContent)
	r.Patch("/api/adr/{number}/status", s.handleUpdateStatus)
	r.Post("/api/adr/{number}/relations", s.handleAddRelation)

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

func (s *Server) handleAddRelation(w http.ResponseWriter, r *http.Request) {
	if s.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}

	if s.relator == nil {
		http.Error(w, "relations not supported", http.StatusNotImplemented)
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
		RelatedTo int `json:"relatedTo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.RelatedTo <= 0 {
		http.Error(w, "relatedTo must be a positive integer", http.StatusBadRequest)
		return
	}

	if body.RelatedTo == number {
		http.Error(w, "cannot relate an ADR to itself", http.StatusBadRequest)
		return
	}

	record, err := s.relator.AddRelation(r.Context(), number, body.RelatedTo)
	if err != nil {
		if errors.Is(err, adr.ErrNotFound) {
			http.Error(w, "ADR not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to add relation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(toDetailResponse(*record)); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if s.config == nil {
		http.Error(w, "config not available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"template": s.config.Template}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleGetTemplateSections(w http.ResponseWriter, r *http.Request) {
	if s.config == nil {
		http.Error(w, "config not available", http.StatusServiceUnavailable)
		return
	}

	sections, err := adr.TemplateSections(s.config.Template)
	if err != nil {
		http.Error(w, "failed to load template sections", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sections); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleGetScopes(w http.ResponseWriter, r *http.Request) {
	if s.scopeStore == nil {
		http.Error(w, "scopes not available", http.StatusServiceUnavailable)
		return
	}

	scopes := s.scopeStore.Scopes()
	if scopes == nil {
		scopes = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(scopes); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleAddScope(w http.ResponseWriter, r *http.Request) {
	if s.scopeStore == nil {
		http.Error(w, "scopes not available", http.StatusServiceUnavailable)
		return
	}

	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	scopes, err := s.scopeStore.AddScope(body.Value)
	if err != nil {
		if errors.Is(err, adr.ErrInvalidScope) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to add scope", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(scopes); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleCreateADR(w http.ResponseWriter, r *http.Request) {
	if s.repo == nil {
		http.Error(w, "repository not configured", http.StatusServiceUnavailable)
		return
	}
	if s.config == nil {
		http.Error(w, "config not available", http.StatusServiceUnavailable)
		return
	}

	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 65536)
	var body struct {
		Title    string            `json:"title"`
		Sections map[string]string `json:"sections,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(body.Title)
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	templateContent, err := adr.TemplateContent(s.config.Template)
	if err != nil {
		http.Error(w, "failed to load template", http.StatusInternalServerError)
		return
	}

	nextNum, err := s.repo.NextNumber(r.Context())
	if err != nil {
		http.Error(w, "failed to determine next number", http.StatusInternalServerError)
		return
	}

	record := adr.New(nextNum, title)
	record.Content = adr.RenderTemplate(templateContent, record)

	// Replace section content with user-provided values
	if len(body.Sections) > 0 {
		sectionDefs, _ := adr.TemplateSections(s.config.Template)
		for _, def := range sectionDefs {
			text, ok := body.Sections[def.Key]
			if !ok || strings.TrimSpace(text) == "" {
				continue
			}
			// "meta" kinds are title-block lines (e.g. Scope); everything else
			// is a "## Heading" body section.
			if def.Kind == "meta" {
				if replaced, found := adr.ReplaceMetaField(record.Content, def.Heading, text); found {
					record.Content = replaced
				}
			} else if replaced, found := adr.ReplaceSectionContent(record.Content, def.Heading, text); found {
				record.Content = replaced
			}
		}
	}

	if err := s.repo.Save(r.Context(), record); err != nil {
		if errors.Is(err, adr.ErrConflict) {
			http.Error(w, "ADR already exists", http.StatusConflict)
			return
		}
		http.Error(w, "failed to save ADR", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/api/adr/%d", record.Number))
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(toDetailResponse(*record)); err != nil {
		log.Printf("error encoding create response: %v", err)
	}
}

func (s *Server) handleUpdateContent(w http.ResponseWriter, r *http.Request) {
	if s.contentUpdater == nil {
		http.Error(w, "content updates not supported", http.StatusNotImplemented)
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

	r.Body = http.MaxBytesReader(w, r.Body, 65536)
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(body.Content) == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}

	// Validate heading number matches URL number
	meta := adr.ExtractMetadata(body.Content)
	if meta.Number > 0 && meta.Number != number {
		http.Error(w, "heading number does not match URL", http.StatusBadRequest)
		return
	}

	record, err := s.contentUpdater.UpdateContent(r.Context(), number, body.Content)
	if err != nil {
		if errors.Is(err, adr.ErrNotFound) {
			http.Error(w, "ADR not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update content", http.StatusInternalServerError)
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
	if _, err := io.Copy(w, f.(io.Reader)); err != nil {
		log.Printf("error serving index.html: %v", err)
	}
}
