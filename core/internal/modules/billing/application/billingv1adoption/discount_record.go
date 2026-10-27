package billingv1adoptionapp

import "time"

type DiscountRecord struct {
	SourceID               string     `json:"source_id"`
	CreatedBy              string     `json:"created_by"`
	ID                     string     `json:"id"`
	VersionID              string     `json:"version_id"`
	PublicationOperationID string     `json:"publication_operation_id"`
	NormalizedCode         string     `json:"normalized_code"`
	Name                   string     `json:"name"`
	Description            string     `json:"description"`
	DiscountKind           string     `json:"discount_kind"`
	FixedMinorUnits        *int64     `json:"fixed_minor_units"`
	Currency               *string    `json:"currency"`
	BasisPoints            *int32     `json:"basis_points"`
	Duration               string     `json:"duration"`
	RepeatingCycles        *int64     `json:"repeating_cycles"`
	Priority               int32      `json:"priority"`
	GlobalUseLimit         *int64     `json:"global_use_limit"`
	PerTenantUseLimit      *int64     `json:"per_tenant_use_limit"`
	SourceUsedCount        int64      `json:"source_used_count"`
	AllOfferings           bool       `json:"all_offerings"`
	OfferingVersionIDs     []string   `json:"offering_version_ids"`
	ApplicabilityEvidence  *string    `json:"applicability_evidence_sha256"`
	ValidStart             time.Time  `json:"valid_start"`
	ValidEnd               *time.Time `json:"valid_end"`
	IsActive               bool       `json:"is_active"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	RevokedAt              *time.Time `json:"revoked_at"`
}
