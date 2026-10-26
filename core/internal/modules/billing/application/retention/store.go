package retention

import (
	"context"

	"github.com/google/uuid"
)

type Store interface {
	Execute(context.Context, Claim, uuid.UUID) (Outcome, error)
}

type IDSource func() (uuid.UUID, error)
