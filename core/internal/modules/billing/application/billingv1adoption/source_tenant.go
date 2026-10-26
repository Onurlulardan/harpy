package billingv1adoptionapp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SourceSubscription struct {
	ID                  uuid.UUID
	TenantID            uuid.UUID
	PlanID              uuid.UUID
	State               string
	PeriodStart         time.Time
	PeriodEnd           time.Time
	TrialStart          *time.Time
	TrialEnd            *time.Time
	CancelAtPeriodEnd   bool
	CancelledAt         *time.Time
	CancellationReason  string
	EffectivePriceCents int64
	CreatedAt           *time.Time
	UpdatedAt           *time.Time
	DeletedAt           *time.Time
}

type SourceRedemption struct {
	ID                  uuid.UUID
	SubscriptionID      uuid.UUID
	CouponID            uuid.UUID
	TenantID            uuid.UUID
	DiscountAmountCents int64
	RemainingMonths     *int64
	AppliedAt           time.Time
	CreatedAt           *time.Time
}

type SourceUsage struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Month     string
	Count     int64
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type GaugeOpening struct {
	EntitlementKey string
	Value          int64
	OwnerFactID    uuid.UUID
	EvidenceSHA256 string
	MeasuredAt     time.Time
}

type SourceBillingEvent struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	SubscriptionID *uuid.UUID
	EventType      string
	AmountCents    int64
	Currency       string
	Description    string
	Metadata       json.RawMessage
	CreatedAt      time.Time
}

type TenantSnapshot struct {
	TenantID      uuid.UUID
	SnapshotAt    time.Time
	Subscriptions []SourceSubscription
	Redemptions   []SourceRedemption
	Usage         []SourceUsage
	Gauges        []GaugeOpening
	Events        []SourceBillingEvent
}

type AdoptionSnapshot struct {
	Global  GlobalSnapshot
	Tenants []TenantSnapshot
}
