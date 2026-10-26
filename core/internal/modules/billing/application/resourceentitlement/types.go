package resourceentitlement

import (
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	"github.com/google/uuid"
)

type CountedState string

const (
	CountedActive   CountedState = "active"
	CountedReleased CountedState = "released"
)

func (state CountedState) Valid() bool {
	return state == CountedActive || state == CountedReleased
}

type ResourceProjection struct {
	Resource    billingapi.MeteredResourceKind
	OwnerFactID uuid.UUID
	State       CountedState
	Revision    int64
	UpdatedAt   time.Time
}

type StoredMutation struct {
	Decision   billingapi.MeteredResourceMutationDecision
	RecordedAt time.Time
}

type Inspection struct {
	Stored     *StoredMutation
	Projection *ResourceProjection
}

type ProjectionChange struct {
	Before *ResourceProjection
	After  *ResourceProjection
}

type MutationRecord struct {
	Decision        billingapi.MeteredResourceMutationDecision
	Change          *ProjectionChange
	UsageDecisionID uuid.UUID
	UsageEntryID    uuid.UUID
	RecordedAt      time.Time
}

type UsageLink struct {
	DecisionID uuid.UUID
	EntryID    uuid.UUID
}

type GaugeSnapshot struct {
	Definition entitlement.Definition
	Limit      entitlement.Value
	PeriodID   uuid.UUID
	Current    int64
}

type GaugeAdjustmentRequest struct {
	EntitlementKey entitlement.Key
	PeriodID       uuid.UUID
	OwnerFactID    uuid.UUID
	Delta          int64
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
}

type GaugeAdjustment struct {
	DecisionID uuid.UUID
	Entry      GaugeEntry
	Allowed    bool
	Replay     bool
	Before     int64
	After      int64
	Limit      entitlement.Value
}

type StoredReconciliation struct {
	Decision   billingapi.MeteredResourceReconcileDecision
	RecordedAt time.Time
}

type ReconciliationInspection struct {
	Stored     *StoredReconciliation
	Projection *ResourceProjection
}

type ReconciliationRecord struct {
	Decision        billingapi.MeteredResourceReconcileDecision
	Evidence        billingapi.MeteredResourceEvidence
	Change          *ProjectionChange
	UsageDecisionID uuid.UUID
	UsageEntryID    uuid.UUID
	RecordedAt      time.Time
}
