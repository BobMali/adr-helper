package adr

import (
	"strings"
	"time"
)

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

// AllStatuses returns all valid ADR statuses.
func AllStatuses() []Status {
	return []Status{Proposed, Accepted, Rejected, Deprecated, Superseded}
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
	Number int
	Title  string
	Status Status
	Date   time.Time
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
