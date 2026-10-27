package billingv1adoptionapp

import "time"

type PriceRecord struct {
	ID         string `json:"id"`
	Cadence    string `json:"cadence"`
	MinorUnits int64  `json:"minor_units"`
	Currency   string `json:"currency"`
}

type GrantRecord struct {
	EntitlementKey string `json:"entitlement_key"`
	ValueType      string `json:"value_type"`
	BooleanValue   *bool  `json:"boolean_value"`
	IntegerValue   *int64 `json:"integer_value"`
	Unlimited      bool   `json:"unlimited"`
}

type OfferingRecord struct {
	SourceID               string        `json:"source_id"`
	ID                     string        `json:"id"`
	VersionID              string        `json:"version_id"`
	PublicationOperationID string        `json:"publication_operation_id"`
	Key                    string        `json:"key"`
	Name                   string        `json:"name"`
	Description            string        `json:"description"`
	SortOrder              int32         `json:"sort_order"`
	HighlightLabel         string        `json:"highlight_label"`
	Color                  string        `json:"color"`
	BillingInterval        string        `json:"billing_interval"`
	IsBase                 bool          `json:"is_base"`
	IsPublic               bool          `json:"is_public"`
	Available              bool          `json:"available"`
	TrialDays              int32         `json:"trial_days"`
	CreatedAt              time.Time     `json:"created_at"`
	UpdatedAt              time.Time     `json:"updated_at"`
	RetiredAt              *time.Time    `json:"retired_at"`
	Prices                 []PriceRecord `json:"prices"`
	Grants                 []GrantRecord `json:"grants"`
}

type CampaignRecord struct {
	SourceID               string     `json:"source_id"`
	CreatedBy              string     `json:"created_by"`
	ID                     string     `json:"id"`
	VersionID              string     `json:"version_id"`
	PublicationOperationID string     `json:"publication_operation_id"`
	Key                    string     `json:"key"`
	Name                   string     `json:"name"`
	Description            string     `json:"description"`
	BannerText             string     `json:"banner_text"`
	AutoApply              bool       `json:"auto_apply"`
	Priority               int32      `json:"priority"`
	IsActive               bool       `json:"is_active"`
	StartsAt               time.Time  `json:"starts_at"`
	EndsAt                 *time.Time `json:"ends_at"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	RetiredAt              *time.Time `json:"retired_at"`
	DiscountVersionIDs     []string   `json:"discount_version_ids"`
}
