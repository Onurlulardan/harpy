package subscriptionapp

import (
	"time"

	"github.com/google/uuid"
)

type ChangeState string

const (
	ChangeScheduled ChangeState = "scheduled"
	ChangeApplied   ChangeState = "applied"
)

type SubscriptionChange struct {
	ID                      uuid.UUID
	DecisionID              uuid.UUID
	SubscriptionID          uuid.UUID
	BeforeOfferingVersionID uuid.UUID
	BeforePriceVersionID    uuid.UUID
	AfterOfferingVersionID  uuid.UUID
	AfterPriceVersionID     uuid.UUID
	SourceOperationID       uuid.UUID
	State                   ChangeState
	EffectiveAt             time.Time
	CreatedAt               time.Time
	Revision                int64
}

func (change SubscriptionChange) valid() bool {
	return validV7(change.ID) && validV7(change.DecisionID) &&
		validID(change.SubscriptionID) && validID(change.BeforeOfferingVersionID) &&
		validID(change.BeforePriceVersionID) && validID(change.AfterOfferingVersionID) &&
		validID(change.AfterPriceVersionID) && validV7(change.SourceOperationID) &&
		change.State == ChangeScheduled && validInstant(change.EffectiveAt) &&
		validInstant(change.CreatedAt) && !change.EffectiveAt.Before(change.CreatedAt) &&
		change.Revision == 1
}

type ChangeResolution struct {
	ID               uuid.UUID
	ExpectedRevision int64
	State            ChangeState
	ResolvedAt       time.Time
}

func (resolution ChangeResolution) valid() bool {
	return validV7(resolution.ID) && resolution.ExpectedRevision > 0 &&
		resolution.State == ChangeApplied && validInstant(resolution.ResolvedAt)
}
