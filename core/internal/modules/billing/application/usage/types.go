package usage

import (
	"time"

	"github.com/google/uuid"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

type ReservationSnapshot struct {
	Definition                 entitlement.Definition
	Limit                      entitlement.Value
	PeriodID                   uuid.UUID
	Period                     billingperiod.Period
	Committed, Reserved, Gauge int64
	Revision                   int64
}

type Reservation struct {
	ID                uuid.UUID
	DecisionID        uuid.UUID
	EntitlementKey    entitlement.Key
	PeriodID          uuid.UUID
	SourceOperationID uuid.UUID
	Amount            int64
	State             usagedomain.ReservationState
	ExpiresAt         time.Time
	EffectStartedAt   *time.Time
	EffectReceiptID   *uuid.UUID
	ResolutionReceipt *uuid.UUID
	CommittedAmount   int64
	Revision          int64
}

type ReservationReplay struct {
	Reservation Reservation
	Decision    DecisionRecord
}

type DecisionRecord struct {
	ID                uuid.UUID
	EntitlementKey    entitlement.Key
	PeriodID          uuid.UUID
	SourceOperationID uuid.UUID
	Current           int64
	Requested         int64
	Limit             entitlement.Value
	Allowed           bool
	SoftExceeded      bool
	DecidedAt         time.Time
}

type ReserveRequest struct {
	EntitlementKey       entitlement.Key
	PeriodID             uuid.UUID
	SourceOperationID    uuid.UUID
	ReservationID        uuid.UUID
	Amount               int64
	ExpiresAt            time.Time
	ResolveCurrentPeriod bool
}

type ReserveResult struct {
	Reservation       Reservation
	Allowed           bool
	SoftLimitExceeded bool
	Replay            bool
	Current           int64
	Requested         int64
	Limit             entitlement.Value
	Period            billingperiod.Period
}

type StartEffectRequest struct {
	ReservationID uuid.UUID
	ReceiptID     uuid.UUID
}

type TransitionRequest struct {
	ReservationID   uuid.UUID
	ReceiptID       *uuid.UUID
	CommittedAmount int64
}

type ExpireRequest struct {
	ReservationID uuid.UUID
	Evidence      usagedomain.EffectEvidence
	ReceiptID     *uuid.UUID
}

type ReservationResult struct {
	Reservation Reservation
	Replay      bool
}

type GaugeSnapshot struct {
	Definition entitlement.Definition
	Limit      entitlement.Value
	PeriodID   uuid.UUID
	Period     billingperiod.Period
	Current    int64
	Revision   int64
}

type GaugeEntry struct {
	ID             uuid.UUID
	DecisionID     uuid.UUID
	EntitlementKey entitlement.Key
	PeriodID       uuid.UUID
	OwnerFactID    uuid.UUID
	Delta          int64
	Before         int64
	After          int64
	CreatedAt      time.Time
}

type GaugeReplay struct {
	Entry    GaugeEntry
	Decision DecisionRecord
}

type AdjustGaugeRequest struct {
	EntitlementKey entitlement.Key
	PeriodID       uuid.UUID
	OwnerFactID    uuid.UUID
	Delta          int64
}

type AdjustGaugeResult struct {
	DecisionID        uuid.UUID
	Entry             GaugeEntry
	Allowed           bool
	SoftLimitExceeded bool
	Replay            bool
	Before            int64
	After             int64
	Limit             entitlement.Value
	Period            billingperiod.Period
}
