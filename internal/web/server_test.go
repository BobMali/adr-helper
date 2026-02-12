package web_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
	adrs   []adr.ADR
	err    error
	getADR *adr.ADR
	getErr error
}

func (m *mockRepo) List(_ context.Context) ([]adr.ADR, error) { return m.adrs, m.err }
func (m *mockRepo) Get(_ context.Context, _ int) (*adr.ADR, error) {
	if m.getADR != nil || m.getErr != nil {
		return m.getADR, m.getErr
	}
	return nil, fmt.Errorf("not implemented")
}
func (m *mockRepo) Save(_ context.Context, _ *adr.ADR) error  { return nil }
func (m *mockRepo) NextNumber(_ context.Context) (int, error) { return 0, nil }

var _ web.StatusUpdater = (*mockUpdater)(nil)

type mockUpdater struct {
	result *adr.ADR
	err    error
	called bool
}

func (m *mockUpdater) UpdateStatus(_ context.Context, _ int, _ string) (*adr.ADR, error) {
	m.called = true
	return m.result, m.err
}

var _ web.Superseder = (*mockSuperseder)(nil)

type mockSuperseder struct {
	result     *adr.ADR
	err        error
	calledWith [2]int // [supersededNum, supersedingNum]
	called     bool
}

func (m *mockSuperseder) Supersede(_ context.Context, supersededNum, supersedingNum int) (*adr.ADR, error) {
	m.called = true
	m.calledWith = [2]int{supersededNum, supersedingNum}
	return m.result, m.err
}

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

// --- GET /api/adr/{number} ---

func TestGetADR_ReturnsADRWithContent(t *testing.T) {
	repo := &mockRepo{
		getADR: &adr.ADR{
			Number:  1,
			Title:   "Use Go",
			Status:  adr.Accepted,
			Date:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Content: "# 1. Use Go\n\n## Status\n\nAccepted\n",
		},
	}
	srv := web.NewServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/adr/1", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, float64(1), body["number"])
	assert.Equal(t, "Use Go", body["title"])
	assert.Equal(t, "Accepted", body["status"])
	assert.Equal(t, "2024-01-15", body["date"])
	assert.Equal(t, "# 1. Use Go\n\n## Status\n\nAccepted\n", body["content"])
}

func TestGetADR_NotFound(t *testing.T) {
	repo := &mockRepo{
		getErr: fmt.Errorf("ADR 0099: %w", adr.ErrNotFound),
	}
	srv := web.NewServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/adr/99", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetADR_InvalidNumber(t *testing.T) {
	repo := &mockRepo{}
	srv := web.NewServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/adr/abc", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetADR_NilRepo(t *testing.T) {
	srv := web.NewServer(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/adr/1", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

// --- GET /api/adr/statuses ---

func TestStatuses_ReturnsAllStatuses(t *testing.T) {
	srv := web.NewServer(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/adr/statuses", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var statuses []string
	err := json.Unmarshal(rec.Body.Bytes(), &statuses)
	require.NoError(t, err)
	assert.Equal(t, []string{"Proposed", "Accepted", "Rejected", "Deprecated", "Superseded"}, statuses)
}

// --- PATCH /api/adr/{number}/status ---

func TestUpdateStatus_Success(t *testing.T) {
	updater := &mockUpdater{
		result: &adr.ADR{
			Number:  1,
			Title:   "Use Go",
			Status:  adr.Accepted,
			Date:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Content: "# 1. Use Go\n\n## Status\n\nAccepted\n",
		},
	}
	repo := &mockRepo{}
	srv := web.NewServer(repo, web.WithStatusUpdater(updater))

	body := strings.NewReader(`{"status":"accepted"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/adr/1/status", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["number"])
	assert.Equal(t, "Accepted", resp["status"])
	assert.NotEmpty(t, resp["content"])
}

func TestUpdateStatus_InvalidNumber(t *testing.T) {
	repo := &mockRepo{}
	updater := &mockUpdater{}
	srv := web.NewServer(repo, web.WithStatusUpdater(updater))

	body := strings.NewReader(`{"status":"accepted"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/adr/abc/status", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateStatus_InvalidStatus(t *testing.T) {
	repo := &mockRepo{}
	updater := &mockUpdater{}
	srv := web.NewServer(repo, web.WithStatusUpdater(updater))

	body := strings.NewReader(`{"status":"invalid"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/adr/1/status", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateStatus_NotFound(t *testing.T) {
	updater := &mockUpdater{
		err: fmt.Errorf("ADR 0099: %w", adr.ErrNotFound),
	}
	repo := &mockRepo{}
	srv := web.NewServer(repo, web.WithStatusUpdater(updater))

	body := strings.NewReader(`{"status":"accepted"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/adr/99/status", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateStatus_NoBody(t *testing.T) {
	repo := &mockRepo{}
	updater := &mockUpdater{}
	srv := web.NewServer(repo, web.WithStatusUpdater(updater))

	req := httptest.NewRequest(http.MethodPatch, "/api/adr/1/status", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateStatus_NoUpdater(t *testing.T) {
	repo := &mockRepo{}
	srv := web.NewServer(repo)

	body := strings.NewReader(`{"status":"accepted"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/adr/1/status", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotImplemented, rec.Code)
}

func TestUpdateStatus_NilRepo(t *testing.T) {
	srv := web.NewServer(nil)

	body := strings.NewReader(`{"status":"accepted"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/adr/1/status", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

// --- SPA handler tests ---

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

// --- PATCH /api/adr/{number}/status — Supersede flow ---

func TestUpdateStatus_Supersede_Success(t *testing.T) {
	superseder := &mockSuperseder{
		result: &adr.ADR{
			Number:  2,
			Title:   "Use Chi",
			Status:  adr.Superseded,
			Date:    time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			Content: "# 2. Use Chi\n\n## Status\n\nSuperseded by [ADR-0003](0003-use-gin.md)\n",
		},
	}
	updater := &mockUpdater{}
	repo := &mockRepo{}
	srv := web.NewServer(repo, web.WithStatusUpdater(updater), web.WithSuperseder(superseder))

	body := strings.NewReader(`{"status":"Superseded","supersededBy":3}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/adr/2/status", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, superseder.called)
	assert.Equal(t, [2]int{2, 3}, superseder.calledWith)
	assert.False(t, updater.called, "UpdateStatus should not be called for Superseded")

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Superseded", resp["status"])
}

func TestUpdateStatus_Supersede_MissingSupersededBy(t *testing.T) {
	superseder := &mockSuperseder{}
	updater := &mockUpdater{}
	repo := &mockRepo{}
	srv := web.NewServer(repo, web.WithStatusUpdater(updater), web.WithSuperseder(superseder))

	body := strings.NewReader(`{"status":"Superseded"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/adr/2/status", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "supersededBy is required")
	assert.False(t, superseder.called)
}

func TestUpdateStatus_Supersede_NoSupersederConfigured(t *testing.T) {
	updater := &mockUpdater{}
	repo := &mockRepo{}
	srv := web.NewServer(repo, web.WithStatusUpdater(updater))

	body := strings.NewReader(`{"status":"Superseded","supersededBy":3}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/adr/2/status", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotImplemented, rec.Code)
	assert.Contains(t, rec.Body.String(), "supersede not supported")
}

func TestUpdateStatus_NonSuperseded_StillUsesUpdater(t *testing.T) {
	superseder := &mockSuperseder{}
	updater := &mockUpdater{
		result: &adr.ADR{
			Number:  1,
			Title:   "Use Go",
			Status:  adr.Accepted,
			Date:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Content: "# 1. Use Go\n\n## Status\n\nAccepted\n",
		},
	}
	repo := &mockRepo{}
	srv := web.NewServer(repo, web.WithStatusUpdater(updater), web.WithSuperseder(superseder))

	body := strings.NewReader(`{"status":"Accepted"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/adr/1/status", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, updater.called, "UpdateStatus should be called for non-Superseded")
	assert.False(t, superseder.called, "Superseder should not be called for non-Superseded")
}

// --- GET /api/adr?q= (filter) ---

func TestListADRs_FilterByTitleQuery(t *testing.T) {
	repo := &mockRepo{adrs: []adr.ADR{
		{Number: 1, Title: "Use Go", Status: adr.Accepted},
		{Number: 2, Title: "Use Chi Router", Status: adr.Proposed},
		{Number: 3, Title: "Use PostgreSQL", Status: adr.Accepted},
	}}
	srv := web.NewServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/adr?q=chi", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body []map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err)
	require.Len(t, body, 1)
	assert.Equal(t, "Use Chi Router", body[0]["title"])
}

func TestListADRs_FilterByNumber(t *testing.T) {
	repo := &mockRepo{adrs: []adr.ADR{
		{Number: 12, Title: "Use PostgreSQL", Status: adr.Accepted},
		{Number: 120, Title: "Use Redis", Status: adr.Proposed},
	}}
	srv := web.NewServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/adr?q=12", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body []map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err)
	require.Len(t, body, 1)
	assert.Equal(t, float64(12), body[0]["number"])
}

func TestListADRs_EmptyQueryReturnsAll(t *testing.T) {
	repo := &mockRepo{adrs: []adr.ADR{
		{Number: 1, Title: "Use Go", Status: adr.Accepted},
		{Number: 2, Title: "Use Chi", Status: adr.Proposed},
	}}
	srv := web.NewServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/adr?q=", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body []map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Len(t, body, 2)
}

func TestListADRs_FilterNoMatches_ReturnsEmptyArray(t *testing.T) {
	repo := &mockRepo{adrs: []adr.ADR{
		{Number: 1, Title: "Use Go", Status: adr.Accepted},
	}}
	srv := web.NewServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/adr?q=zzz", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "[]", trimNewline(rec.Body.String()))
}

func trimNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}
