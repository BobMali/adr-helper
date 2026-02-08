package adr_test

import (
	"context"
	"testing"
	"time"

	"github.com/malek/adr-helper/internal/adr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	before := time.Now()
	record := adr.New(1, "Use Go for CLI")
	after := time.Now()

	assert.Equal(t, 1, record.Number)
	assert.Equal(t, "Use Go for CLI", record.Title)
	assert.Equal(t, adr.Proposed, record.Status)
	require.False(t, record.Date.IsZero(), "Date should be set")
	assert.False(t, record.Date.Before(before), "Date should not be before test start")
	assert.False(t, record.Date.After(after), "Date should not be after test end")
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status   adr.Status
		expected string
	}{
		{adr.Proposed, "Proposed"},
		{adr.Accepted, "Accepted"},
		{adr.Deprecated, "Deprecated"},
		{adr.Superseded, "Superseded"},
		{adr.Status(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

// Compile-time interface check
var _ adr.Repository = (*mockRepo)(nil)

type mockRepo struct{}

func (m *mockRepo) List(_ context.Context) ([]adr.ADR, error)        { return nil, nil }
func (m *mockRepo) Get(_ context.Context, _ int) (*adr.ADR, error)   { return nil, nil }
func (m *mockRepo) Save(_ context.Context, _ *adr.ADR) error         { return nil }
func (m *mockRepo) NextNumber(_ context.Context) (int, error)        { return 0, nil }
