package billingv1

import "encoding/json"

type planRequest struct {
	Name             *string          `json:"name"`
	Slug             *string          `json:"slug"`
	Description      *string          `json:"description"`
	PriceCents       *int64           `json:"price_cents"`
	AnnualPriceCents *int64           `json:"annual_price_cents"`
	BillingInterval  *string          `json:"billing_interval"`
	Currency         *string          `json:"currency"`
	Limits           *json.RawMessage `json:"limits"`
	Features         *json.RawMessage `json:"features"`
	SortOrder        *int             `json:"sort_order"`
	IsActive         *bool            `json:"is_active"`
	IsPublic         *bool            `json:"is_public"`
	HighlightLabel   *string          `json:"highlight_label"`
	Color            *string          `json:"color"`
	TrialDays        *int             `json:"trial_days"`
}

type planLimits struct {
	SubmissionsPerMonth int64 `json:"submissions_per_month"`
	MaxUsers            int64 `json:"max_users"`
	MaxForms            int64 `json:"max_forms"`
	MaxStorageMB        int64 `json:"max_storage_mb"`
}

type planFeatures struct {
	AIChat    bool `json:"ai_chat"`
	Workflows bool `json:"workflows"`
	APIAccess bool `json:"api_access"`
}
