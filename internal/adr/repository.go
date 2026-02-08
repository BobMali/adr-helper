package adr

import "context"

// Repository defines the persistence operations for ADRs.
type Repository interface {
	List(ctx context.Context) ([]ADR, error)
	Get(ctx context.Context, number int) (*ADR, error)
	Save(ctx context.Context, record *ADR) error
	NextNumber(ctx context.Context) (int, error)
}
