package subscriptionapp

import (
	"time"

	"github.com/google/uuid"
)

type BoundaryKind string

const (
	BoundaryCancellation BoundaryKind = "cancellation"
	BoundaryManualChange BoundaryKind = "manual_change"
	BoundaryTrialEnd     BoundaryKind = "trial_end"
	BoundaryGraceEnd     BoundaryKind = "grace_end"
	BoundaryPeriodEnd    BoundaryKind = "period_end"
)

func (kind BoundaryKind) valid() bool {
	switch kind {
	case BoundaryCancellation, BoundaryManualChange, BoundaryTrialEnd,
		BoundaryGraceEnd, BoundaryPeriodEnd:
		return true
	default:
		return false
	}
}

type JobIntent struct {
	ID               uuid.UUID
	EffectID         uuid.UUID
	SubscriptionID   uuid.UUID
	Boundary         BoundaryKind
	RunAt            time.Time
	ExpectedRevision int64
	PayloadRevision  int
}

func (intent JobIntent) valid() bool {
	return validV7(intent.ID) && validV7(intent.EffectID) &&
		validID(intent.SubscriptionID) && intent.Boundary.valid() &&
		validInstant(intent.RunAt) && intent.ExpectedRevision > 0 &&
		intent.PayloadRevision == 1
}

type JobCancellation struct {
	ID          uuid.UUID
	CancelledAt time.Time
	Reason      string
}

func (value JobCancellation) valid() bool {
	return validV7(value.ID) && validInstant(value.CancelledAt) && value.Reason != ""
}

type JobCompletion struct {
	ID              uuid.UUID
	LeaseGeneration int64
	CompletedAt     time.Time
}

func (value JobCompletion) valid() bool {
	return validV7(value.ID) && value.LeaseGeneration > 0 && validInstant(value.CompletedAt)
}
