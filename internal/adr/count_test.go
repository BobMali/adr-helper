package adr

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountByStatus_EmptySlice(t *testing.T) {
	counts := CountByStatus(nil)

	assert.Equal(t, 0, counts.Total)
	for _, s := range AllStatuses() {
		assert.Equal(t, 0, counts.ByStatus[s], "expected 0 for %s", s)
	}
}

func TestCountByStatus_MixedStatuses(t *testing.T) {
	records := []ADR{
		{Number: 1, Status: Accepted},
		{Number: 2, Status: Accepted},
		{Number: 3, Status: Proposed},
		{Number: 4, Status: Rejected},
	}

	counts := CountByStatus(records)

	assert.Equal(t, 4, counts.Total)
	assert.Equal(t, 1, counts.ByStatus[Proposed])
	assert.Equal(t, 2, counts.ByStatus[Accepted])
	assert.Equal(t, 1, counts.ByStatus[Rejected])
	assert.Equal(t, 0, counts.ByStatus[Deprecated])
	assert.Equal(t, 0, counts.ByStatus[Superseded])
}

func TestCountByStatus_AllStatusesPresent(t *testing.T) {
	// Even with an empty slice, the map should contain all 5 statuses.
	counts := CountByStatus([]ADR{})

	assert.Len(t, counts.ByStatus, len(AllStatuses()))
	for _, s := range AllStatuses() {
		_, ok := counts.ByStatus[s]
		assert.True(t, ok, "expected key for %s", s)
	}
}

func TestStatusCounts_JSONMarshal(t *testing.T) {
	counts := CountByStatus([]ADR{
		{Number: 1, Status: Accepted},
		{Number: 2, Status: Proposed},
	})

	data, err := json.Marshal(counts)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))

	// Top-level keys
	assert.Contains(t, raw, "byStatus")
	assert.Contains(t, raw, "total")

	// byStatus should use string keys
	var byStatus map[string]int
	require.NoError(t, json.Unmarshal(raw["byStatus"], &byStatus))
	assert.Equal(t, 1, byStatus["Accepted"])
	assert.Equal(t, 1, byStatus["Proposed"])
	assert.Equal(t, 0, byStatus["Rejected"])

	var total int
	require.NoError(t, json.Unmarshal(raw["total"], &total))
	assert.Equal(t, 2, total)
}
