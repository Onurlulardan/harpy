package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type analyticsJobStore interface {
	WarmQuery(context.Context, uuid.UUID, uuid.UUID, [32]byte) error
	VerifyRevalidated(context.Context, uuid.UUID, uuid.UUID, [32]byte) error
	ApplyPublicationRepair(context.Context, uuid.UUID, uint64, time.Time) error
}
