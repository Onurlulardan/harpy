package retention

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Cycle struct {
	Claim          Claim
	CompletedAt    time.Time
	SuccessorJobID uuid.UUID
}

type Store interface {
	ApplyCycle(context.Context, Cycle) (Outcome, error)
}

type IDSource func() (uuid.UUID, error)
type Clock func() time.Time
