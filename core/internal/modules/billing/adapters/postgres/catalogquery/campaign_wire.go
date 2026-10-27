package catalogquery

import (
	"encoding/json"
	"fmt"
	"time"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	"github.com/google/uuid"
)

type discountWire struct {
	ID                      uuid.UUID   `json:"id"`
	LegacyID                uuid.UUID   `json:"legacy_id"`
	CreatedBy               uuid.UUID   `json:"created_by"`
	Code                    string      `json:"code"`
	Revision                int64       `json:"revision"`
	VersionID               uuid.UUID   `json:"version_id"`
	Version                 int64       `json:"version"`
	State                   string      `json:"state"`
	Name                    string      `json:"name"`
	Description             string      `json:"description"`
	Available               bool        `json:"available"`
	Kind                    string      `json:"kind"`
	BasisPoints             *int64      `json:"basis_points"`
	FixedMinorUnits         *int64      `json:"fixed_minor_units"`
	Currency                *string     `json:"currency"`
	CurrencyRegistryVersion *int64      `json:"currency_registry_version"`
	Duration                string      `json:"duration"`
	DurationCycles          *int64      `json:"duration_cycles"`
	Priority                int32       `json:"priority"`
	AllOfferings            bool        `json:"all_offerings"`
	OfferingVersionIDs      []uuid.UUID `json:"offering_version_ids"`
	StableOfferingIDs       []uuid.UUID `json:"stable_offering_ids"`
	LegacyOfferingIDs       []uuid.UUID `json:"legacy_offering_ids"`
	UsedCount               int64       `json:"used_count"`
	CampaignIDs             []uuid.UUID `json:"campaign_ids"`
	LegacyCampaignIDs       []uuid.UUID `json:"legacy_campaign_ids"`
	ValidStart              time.Time   `json:"valid_start"`
	ValidEnd                *time.Time  `json:"valid_end"`
	GlobalUseLimit          *int64      `json:"global_use_limit"`
	PerTenantUseLimit       *int64      `json:"per_tenant_use_limit"`
	CreatedAt               time.Time   `json:"created_at"`
	UpdatedAt               time.Time   `json:"updated_at"`
}

func decodeDiscount(document []byte) (catalogapp.DiscountRecord, error) {
	var value discountWire
	if err := json.Unmarshal(document, &value); err != nil {
		return catalogapp.DiscountRecord{}, fmt.Errorf("decode active billing discount: %w", err)
	}
	return catalogapp.DiscountRecord{ID: value.ID, LegacyID: value.LegacyID,
		CreatedBy: value.CreatedBy, Code: value.Code,
		Revision: value.Revision, VersionID: value.VersionID, Version: value.Version,
		State: value.State, Name: value.Name, Description: value.Description,
		Available: value.Available,
		Kind:      value.Kind, BasisPoints: value.BasisPoints,
		FixedMinorUnits: value.FixedMinorUnits, Currency: value.Currency,
		CurrencyRegistryVersion: value.CurrencyRegistryVersion,
		Duration:                value.Duration, DurationCycles: value.DurationCycles,
		Priority: value.Priority, AllOfferings: value.AllOfferings,
		OfferingVersionIDs: value.OfferingVersionIDs,
		StableOfferingIDs:  value.StableOfferingIDs,
		LegacyOfferingIDs:  value.LegacyOfferingIDs,
		UsedCount:          value.UsedCount,
		CampaignIDs:        value.CampaignIDs,
		LegacyCampaignIDs:  value.LegacyCampaignIDs,
		ValidStart:         value.ValidStart, ValidEnd: value.ValidEnd,
		GlobalUseLimit:    value.GlobalUseLimit,
		PerTenantUseLimit: value.PerTenantUseLimit,
		CreatedAt:         value.CreatedAt, UpdatedAt: value.UpdatedAt}, nil
}
