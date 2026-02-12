package adr

// StatusCounts holds ADR counts grouped by status.
type StatusCounts struct {
	ByStatus map[Status]int `json:"byStatus"`
	Total    int            `json:"total"`
}

// CountByStatus tallies ADRs by their status.
// The returned map always contains an entry for every valid status.
func CountByStatus(records []ADR) StatusCounts {
	counts := make(map[Status]int, len(AllStatuses()))
	for _, s := range AllStatuses() {
		counts[s] = 0
	}
	for _, r := range records {
		counts[r.Status]++
	}
	return StatusCounts{ByStatus: counts, Total: len(records)}
}
