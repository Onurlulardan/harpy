package billingv1adoptionapp

import "time"

const ScopePlanContractVersion = "harpy/billing-v1-adoption-plan/v1"

type CategoryEvidence struct {
	Category   string `json:"category"`
	RowCount   int    `json:"row_count"`
	RootSHA256 string `json:"root_sha256"`
}

type ScopePlan struct {
	ContractVersion         string               `json:"contract_version"`
	ScopeKind               string               `json:"scope_kind"`
	ScopeID                 string               `json:"scope_id"`
	SnapshotAt              time.Time            `json:"snapshot_at"`
	SourceTimezone          string               `json:"source_timezone"`
	SourceSHA256            string               `json:"source_sha256"`
	Categories              []CategoryEvidence   `json:"categories"`
	Offerings               []OfferingRecord     `json:"offerings"`
	Campaigns               []CampaignRecord     `json:"campaigns"`
	Discounts               []DiscountRecord     `json:"discounts"`
	ManualProviderAccountID string               `json:"manual_provider_account_id"`
	BaseAssignments         []BaseOfferingRecord `json:"base_assignments"`
	Subscriptions           []SubscriptionRecord `json:"subscriptions"`
	Redemptions             []RedemptionRecord   `json:"redemptions"`
	UsageOpenings           []UsageOpeningRecord `json:"usage_openings"`
	GaugeOpenings           []GaugeOpeningRecord `json:"gauge_openings"`
	LegacyEvents            []LegacyEventRecord  `json:"legacy_events"`
	Issues                  []Issue              `json:"issues"`
	Eligible                bool                 `json:"eligible"`
}

type Bundle struct {
	Global  ScopePlan
	Tenants []ScopePlan
}

func emptyScopePlan(scopeKind, scopeID, timezone string, at time.Time) ScopePlan {
	return ScopePlan{
		ContractVersion: ScopePlanContractVersion,
		ScopeKind:       scopeKind, ScopeID: scopeID, SnapshotAt: at,
		SourceTimezone: timezone,
		Categories:     []CategoryEvidence{}, Offerings: []OfferingRecord{},
		Campaigns: []CampaignRecord{}, Discounts: []DiscountRecord{},
		BaseAssignments: []BaseOfferingRecord{}, Subscriptions: []SubscriptionRecord{},
		Redemptions: []RedemptionRecord{}, UsageOpenings: []UsageOpeningRecord{},
		GaugeOpenings: []GaugeOpeningRecord{}, Issues: []Issue{},
		LegacyEvents: []LegacyEventRecord{},
	}
}
