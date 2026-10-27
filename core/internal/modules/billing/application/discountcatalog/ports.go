package discountcatalog

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Create(context.Context, CreateMutation) (Snapshot, error)
	Publish(context.Context, CreateVersionMutation) (Snapshot, error)
	Retire(context.Context, RetireMutation) (Snapshot, error)
}

type IDGenerator func() (uuid.UUID, error)
type Clock func() time.Time
