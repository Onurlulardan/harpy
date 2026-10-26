package invoice

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Finalize(context.Context, FinalizeExecution, FinalizePlanner) (FinalizeResult, error)
}

type FinalizePlanner func(DraftSnapshot, *Finalization) (FinalizationPlan, error)

type Clock func() time.Time
type IDSource func() (uuid.UUID, error)
