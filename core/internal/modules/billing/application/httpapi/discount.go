package httpapi

type DiscountView struct {
	ID                   string     `json:"id"`
	LegacyID             string     `json:"legacy_id"`
	CreatedBy            string     `json:"created_by"`
	Code                 string     `json:"code"`
	Name                 string     `json:"name"`
	Description          string     `json:"description"`
	Revision             int64      `json:"revision"`
	VersionID            string     `json:"version_id"`
	Version              int64      `json:"version"`
	State                string     `json:"state"`
	Kind                 string     `json:"kind"`
	BasisPoints          *int64     `json:"basis_points,omitempty"`
	FixedAmount          *MoneyView `json:"fixed_amount,omitempty"`
	Duration             string     `json:"duration"`
	DurationCycles       *int       `json:"duration_cycles,omitempty"`
	Priority             int        `json:"priority"`
	AllOfferings         bool       `json:"all_offerings"`
	OfferingIDs          []string   `json:"offering_version_ids"`
	StableOfferingIDs    []string   `json:"offering_ids"`
	LegacyOfferingIDs    []string   `json:"legacy_offering_ids"`
	UsedCount            int64      `json:"used_count"`
	CampaignID           *string    `json:"campaign_id,omitempty"`
	LegacyCampaignID     *string    `json:"legacy_campaign_id,omitempty"`
	AssociationAmbiguous bool       `json:"association_ambiguous,omitempty"`
	Available            bool       `json:"available"`
	ValidStart           string     `json:"valid_start"`
	ValidEnd             *string    `json:"valid_end,omitempty"`
	GlobalUseLimit       *int64     `json:"global_use_limit,omitempty"`
	PerTenantUseLimit    *int64     `json:"per_tenant_use_limit,omitempty"`
	CreatedAt            string     `json:"created_at"`
	UpdatedAt            string     `json:"updated_at"`
}

type ValidateDiscountCommand struct {
	Access            Access
	Code              string
	OfferingVersionID string
}

type RedeemDiscountCommand struct {
	Access         Access
	IdempotencyKey string
	SubscriptionID string
	Code           string
}

type DiscountVersionInput struct {
	Name               string
	Description        string
	Code               string
	Kind               string
	BasisPoints        *int64
	FixedAmount        *MoneyView
	Duration           string
	DurationCycles     *int
	Priority           int
	AllOfferings       bool
	OfferingVersionIDs []string
	ValidStart         string
	ValidEnd           *string
	GlobalUseLimit     *int64
	PerTenantUseLimit  *int64
	Available          *bool
	Campaign           CampaignBindingInput
}

type CampaignBindingInput struct {
	Replace    bool
	CampaignID *string
}

type CreateDiscountCommand struct {
	Access         Access
	IdempotencyKey string
	InitialVersion DiscountVersionInput
}

type CreateDiscountVersionCommand struct {
	Access           Access
	IdempotencyKey   string
	DiscountID       string
	ExpectedRevision int64
	Version          DiscountVersionInput
}

type RetireDiscountCommand struct {
	Access           Access
	IdempotencyKey   string
	DiscountID       string
	ExpectedRevision int64
	Reason           string
}
