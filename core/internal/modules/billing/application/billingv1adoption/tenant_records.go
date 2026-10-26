package billingv1adoptionapp

import "time"

type BaseOfferingRecord struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	OfferingVersionID string    `json:"offering_version_id"`
	SourceOperationID string    `json:"source_operation_id"`
	EffectiveStart    time.Time `json:"effective_start"`
}

type SubscriptionRecord struct {
	SourceID            string     `json:"source_id"`
	ID                  string     `json:"id"`
	TenantID            string     `json:"tenant_id"`
	ProviderAccountID   string     `json:"provider_account_id"`
	OfferingVersionID   string     `json:"offering_version_id"`
	PriceVersionID      string     `json:"price_version_id"`
	State               string     `json:"state"`
	PeriodStart         time.Time  `json:"period_start"`
	PeriodEnd           time.Time  `json:"period_end"`
	TrialStartedAt      *time.Time `json:"trial_started_at"`
	TrialEndsAt         *time.Time `json:"trial_ends_at"`
	GraceEndsAt         *time.Time `json:"grace_ends_at"`
	CancelAtPeriodEnd   bool       `json:"cancel_at_period_end"`
	CancelAt            *time.Time `json:"cancel_at"`
	CancelledAt         *time.Time `json:"cancelled_at"`
	CancellationReason  string     `json:"cancellation_reason"`
	EffectiveMinorUnits int64      `json:"effective_minor_units"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type RedemptionRecord struct {
	SourceID                 string     `json:"source_id"`
	ID                       string     `json:"id"`
	SourceOperationID        string     `json:"source_operation_id"`
	TenantID                 string     `json:"tenant_id"`
	SubscriptionID           string     `json:"subscription_id"`
	DiscountVersionID        string     `json:"discount_version_id"`
	OfferingVersionID        string     `json:"offering_version_id"`
	PriceVersionID           string     `json:"price_version_id"`
	AppliedKind              string     `json:"applied_kind"`
	AppliedFixedMinorUnits   *int64     `json:"applied_fixed_minor_units"`
	AppliedCurrency          *string    `json:"applied_currency"`
	AppliedBasisPoints       *int32     `json:"applied_basis_points"`
	AppliedPriority          int32      `json:"applied_priority"`
	AppliedDuration          string     `json:"applied_duration"`
	AppliedRepeatingCycles   *int64     `json:"applied_repeating_cycles"`
	SourceDiscountMinorUnits int64      `json:"source_discount_minor_units"`
	PolicySHA256             string     `json:"policy_sha256"`
	RedeemedAt               time.Time  `json:"redeemed_at"`
	RelationshipEndedAt      *time.Time `json:"relationship_ended_at"`
}
