package adr

import "time"

// Status represents the lifecycle state of an ADR.
type Status int

const (
	Proposed   Status = iota
	Accepted
	Deprecated
	Superseded
)

func (s Status) String() string {
	switch s {
	case Proposed:
		return "Proposed"
	case Accepted:
		return "Accepted"
	case Deprecated:
		return "Deprecated"
	case Superseded:
		return "Superseded"
	default:
		return "Unknown"
	}
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
