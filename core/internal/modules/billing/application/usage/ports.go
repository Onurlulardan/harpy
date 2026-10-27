package usage

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
	"github.com/google/uuid"
)

type Store interface {
	ResolvePeriod(context.Context, entitlement.Key) (uuid.UUID, error)
	Reserve(context.Context, ReserveExecution, ReservePlanner) (ReserveResult, error)
	Adjust(context.Context, AdjustExecution, AdjustPlanner) (AdjustGaugeResult, error)
	Start(context.Context, StartExecution, ReservationPlanner) (ReservationResult, error)
	Complete(context.Context, TransitionExecution, ReservationPlanner) (ReservationResult, error)
	Release(context.Context, TransitionExecution, ReservationPlanner) (ReservationResult, error)
	Expire(context.Context, ExpireExecution, ReservationPlanner) (ReservationResult, error)
}

type GaugeInspector interface {
	InspectGauge(context.Context, entitlement.Key, uuid.UUID) (GaugeSnapshot, error)
}

type IDGenerator func() (uuid.UUID, error)
type Clock func() time.Time

type EffectStart struct {
	ReservationID    uuid.UUID
	ExpectedRevision int64
	StartedAt        time.Time
	ReceiptID        uuid.UUID
}

type ReservationTransition struct {
	LedgerEntryID    uuid.UUID
	RemainderEntryID uuid.UUID
	ReservationID    uuid.UUID
	ExpectedRevision int64
	From             usagedomain.ReservationState
	To               usagedomain.ReservationState
	ResolvedAt       time.Time
	ReceiptID        *uuid.UUID
	CommittedAmount  int64
}
