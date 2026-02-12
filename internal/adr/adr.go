package adr

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// ErrNotFound is returned when an ADR cannot be found.
var ErrNotFound = errors.New("ADR not found")

// Status represents the lifecycle state of an ADR.
type Status int

const (
	Proposed Status = iota
	Accepted
	Rejected
	Deprecated
	Superseded
)

func (s Status) String() string {
	switch s {
	case Proposed:
		return "Proposed"
	case Accepted:
		return "Accepted"
	case Rejected:
		return "Rejected"
	case Deprecated:
		return "Deprecated"
	case Superseded:
		return "Superseded"
	default:
		return "Unknown"
	}
}

// MarshalJSON encodes Status as a JSON string (e.g. "Accepted").
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// MarshalText implements encoding.TextMarshaler so that map[Status]int
// keys serialize as string names (e.g. "Proposed") in JSON.
func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// AllStatuses returns all valid ADR statuses.
func AllStatuses() []Status {
	return []Status{Proposed, Accepted, Rejected, Deprecated, Superseded}
}

// StatusCategory classifies statuses into color-semantic groups.
type StatusCategory int

const (
	StatusCategoryPending  StatusCategory = iota // Proposed
	StatusCategoryActive                         // Accepted
	StatusCategoryInactive                       // Rejected, Deprecated, Superseded
)

// Category returns the semantic category of the status.
func (s Status) Category() StatusCategory {
	switch s {
	case Accepted:
		return StatusCategoryActive
	case Rejected, Deprecated, Superseded:
		return StatusCategoryInactive
	default:
		return StatusCategoryPending
	}
}

// ParseStatus converts a string to a Status. Returns (status, ok).
// Case-insensitive, also handles prefixes like "superseded by ...".
func ParseStatus(s string) (Status, bool) {
	lower := strings.ToLower(strings.TrimSpace(s))
	if lower == "" {
		return 0, false
	}
	for _, st := range AllStatuses() {
		name := strings.ToLower(st.String())
		if lower == name || strings.HasPrefix(lower, name+" ") {
			return st, true
		}
	}
	return 0, false
}

// AllStatusStrings returns all valid status names in lowercase.
func AllStatusStrings() []string {
	all := AllStatuses()
	result := make([]string, len(all))
	for i, s := range all {
		result[i] = strings.ToLower(s.String())
	}
	return result
}

// ADR represents an Architecture Decision Record.
type ADR struct {
	Number  int
	Title   string
	Status  Status
	Date    time.Time
	Content string
}

// New creates a new ADR with the given number and title, defaulting to Proposed status.
func New(number int, title string) *ADR {
	return &ADR{
		Number: number,
		Title:  title,
		Status: Proposed,
		Date:   time.Now(),
	}
}
