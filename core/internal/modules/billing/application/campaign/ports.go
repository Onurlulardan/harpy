package campaign

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Store commits the catalog mutation, audit/outbox intent, and completed
// idempotency receipt atomically in an owner-bound global transaction.
type Store interface {
	Create(context.Context, CreateMutation) (Snapshot, error)
	Publish(context.Context, CreateVersionMutation) (Snapshot, error)
	Retire(context.Context, RetireMutation) (Snapshot, error)
}

type IDGenerator func() (uuid.UUID, error)
type Clock func() time.Time
