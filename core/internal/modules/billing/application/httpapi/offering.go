package httpapi

import "time"

type MoneyView struct {
	MinorUnits      int64  `json:"minor_units"`
	Currency        string `json:"currency"`
	RegistryVersion int64  `json:"currency_registry_version"`
}

type PriceView struct {
	ID      string    `json:"id"`
	Cadence string    `json:"cadence"`
	Amount  MoneyView `json:"amount"`
	Version int64     `json:"version"`
}

type EntitlementValueView struct {
	Key       string `json:"key"`
	Kind      string `json:"kind"`
	Boolean   *bool  `json:"boolean,omitempty"`
	Integer   *int64 `json:"integer,omitempty"`
	Unlimited bool   `json:"unlimited"`
}

type OfferingView struct {
	ID              string                 `json:"id"`
	Key             string                 `json:"key"`
	Revision        int64                  `json:"revision"`
	VersionID       string                 `json:"version_id"`
	Version         int64                  `json:"version"`
	State           string                 `json:"state"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	SortOrder       int32                  `json:"sort_order"`
	HighlightLabel  string                 `json:"highlight_label,omitempty"`
	Color           string                 `json:"color,omitempty"`
	BillingInterval string                 `json:"billing_interval"`
	Public          bool                   `json:"public"`
	Available       bool                   `json:"available"`
	Base            bool                   `json:"base"`
	TrialDays       int                    `json:"trial_days"`
	GraceDays       int                    `json:"grace_days"`
	EffectiveStart  string                 `json:"effective_start"`
	EffectiveEnd    *string                `json:"effective_end,omitempty"`
	Prices          []PriceView            `json:"prices"`
	Entitlements    []EntitlementValueView `json:"entitlements"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type OfferingVersionInput struct {
	Name            string
	Description     string
	SortOrder       int32
	HighlightLabel  string
	Color           string
	BillingInterval string
	Public          bool
	Available       bool
	Base            bool
	TrialDays       int
	GraceDays       int
	EffectiveStart  string
	EffectiveEnd    *string
	Prices          []PriceInput
	Entitlements    []EntitlementInput
}

type PriceInput struct {
	Cadence         string
	MinorUnits      int64
	Currency        string
	RegistryVersion int64
}

type EntitlementInput struct {
	Key       string
	Kind      string
	Boolean   *bool
	Integer   *int64
	Unlimited bool
}

type CreateOfferingCommand struct {
	Access         Access
	IdempotencyKey string
	Key            string
	InitialVersion OfferingVersionInput
}

type CreateOfferingVersionCommand struct {
	Access           Access
	IdempotencyKey   string
	OfferingID       string
	ExpectedRevision int64
	Version          OfferingVersionInput
}

type RetireOfferingCommand struct {
	Access           Access
	IdempotencyKey   string
	OfferingID       string
	ExpectedRevision int64
	Reason           string
}
