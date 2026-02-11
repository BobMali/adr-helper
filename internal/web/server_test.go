package web_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/BobMali/adr-helper/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ adr.Repository = (*mockRepo)(nil)

type mockRepo struct {
	adrs []adr.ADR
	err  error
}

func (m *mockRepo) List(_ context.Context) ([]adr.ADR, error) { return m.adrs, m.err }
func (m *mockRepo) Get(_ context.Context, _ int) (*adr.ADR, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockRepo) Save(_ context.Context, _ *adr.ADR) error  { return nil }
func (m *mockRepo) NextNumber(_ context.Context) (int, error) { return 0, nil }

func TestHealthEndpoint_ReturnsOK(t *testing.T) {
	srv := web.NewServer(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHealthEndpoint_ReturnsJSON(t *testing.T) {
	srv := web.NewServer(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "ok", body["status"])
}

func TestUnknownRoute_Returns404(t *testing.T) {
	srv := web.NewServer(nil)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListADRs_ReturnsADRs(t *testing.T) {
	repo := &mockRepo{adrs: []adr.ADR{
		{Number: 1, Title: "Use Go", Status: adr.Accepted, Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Number: 2, Title: "Use Chi", Status: adr.Proposed, Date: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
	}}
	srv := web.NewServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body []map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err)
	require.Len(t, body, 2)
	assert.Equal(t, float64(1), body[0]["number"])
	assert.Equal(t, "Use Go", body[0]["title"])
	assert.Equal(t, "Accepted", body[0]["status"])
	assert.Equal(t, "2024-01-15", body[0]["date"])
}

func TestListADRs_EmptyList(t *testing.T) {
	repo := &mockRepo{adrs: []adr.ADR{}}
	srv := web.NewServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "[]", trimNewline(rec.Body.String()))
}

func TestListADRs_RepoError(t *testing.T) {
	repo := &mockRepo{err: fmt.Errorf("disk error")}
	srv := web.NewServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestListADRs_NilRepo(t *testing.T) {
	srv := web.NewServer(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/adr", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestSPAHandler_ServesIndexFallback(t *testing.T) {
	frontend := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>SPA</html>")},
	}
	srv := web.NewServer(nil, web.WithFrontend(frontend))

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Contains(t, rec.Body.String(), "<html>SPA</html>")
}

func TestSPAHandler_ServesStaticAsset(t *testing.T) {
	frontend := fstest.MapFS{
		"index.html":    &fstest.MapFile{Data: []byte("<html>SPA</html>")},
		"assets/app.js": &fstest.MapFile{Data: []byte("console.log('app')")},
	}
	srv := web.NewServer(nil, web.WithFrontend(frontend))

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "public, immutable, max-age=31536000", rec.Header().Get("Cache-Control"))
	assert.Contains(t, rec.Body.String(), "console.log('app')")
}

func TestSPAHandler_ServesRootAsIndex(t *testing.T) {
	frontend := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>SPA</html>")},
	}
	srv := web.NewServer(nil, web.WithFrontend(frontend))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<html>SPA</html>")
}

func trimNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}
