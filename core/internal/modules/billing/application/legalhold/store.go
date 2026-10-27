package legalhold

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Place(context.Context, PlaceMutation) (Snapshot, error)
	Release(context.Context, ReleaseMutation) (Snapshot, error)
}

type IDSource func() (uuid.UUID, error)
type Clock func() time.Time
