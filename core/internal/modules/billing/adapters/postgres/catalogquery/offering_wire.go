package catalogquery

import (
	"encoding/json"
	"fmt"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	"github.com/google/uuid"
)

type priceWire struct {
	ID                      uuid.UUID `json:"id"`
	Cadence                 string    `json:"cadence"`
	MinorUnits              int64     `json:"minor_units"`
	Currency                string    `json:"currency"`
	CurrencyRegistryVersion int64     `json:"currency_registry_version"`
}

type entitlementWire struct {
	Key       string `json:"key"`
	Kind      string `json:"kind"`
	Boolean   *bool  `json:"boolean"`
	Integer   *int64 `json:"integer"`
	Unlimited bool   `json:"unlimited"`
}

func decodeOfferingChildren(
	pricesDocument []byte,
	entitlementsDocument []byte,
) ([]catalogapp.PriceRecord, []catalogapp.EntitlementRecord, error) {
	var prices []priceWire
	var entitlements []entitlementWire
	if err := json.Unmarshal(pricesDocument, &prices); err != nil {
		return nil, nil, fmt.Errorf("decode billing offering prices: %w", err)
	}
	if err := json.Unmarshal(entitlementsDocument, &entitlements); err != nil {
		return nil, nil, fmt.Errorf("decode billing offering entitlements: %w", err)
	}
	priceRecords := make([]catalogapp.PriceRecord, len(prices))
	for index, value := range prices {
		priceRecords[index] = catalogapp.PriceRecord{ID: value.ID, Cadence: value.Cadence,
			MinorUnits: value.MinorUnits, Currency: value.Currency,
			CurrencyRegistryVersion: value.CurrencyRegistryVersion}
	}
	entitlementRecords := make([]catalogapp.EntitlementRecord, len(entitlements))
	for index, value := range entitlements {
		entitlementRecords[index] = catalogapp.EntitlementRecord{Key: value.Key,
			Kind: value.Kind, Boolean: value.Boolean, Integer: value.Integer,
			Unlimited: value.Unlimited}
	}
	return priceRecords, entitlementRecords, nil
}
