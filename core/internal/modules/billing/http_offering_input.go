package billing

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"

type offeringInput struct {
	Key            string               `json:"key"`
	InitialVersion offeringVersionInput `json:"initial_version"`
}

type offeringVersionInput struct {
	ExpectedRevision int64              `json:"expected_revision,omitempty"`
	Name             string             `json:"name"`
	Public           bool               `json:"public"`
	Available        bool               `json:"available"`
	Base             bool               `json:"base"`
	TrialDays        int                `json:"trial_days"`
	GraceDays        int                `json:"grace_days"`
	EffectiveStart   string             `json:"effective_start"`
	EffectiveEnd     *string            `json:"effective_end,omitempty"`
	Prices           []priceInput       `json:"prices"`
	Entitlements     []entitlementInput `json:"entitlements"`
}

type priceInput struct {
	Cadence         string `json:"cadence"`
	MinorUnits      int64  `json:"minor_units"`
	Currency        string `json:"currency"`
	RegistryVersion int64  `json:"currency_registry_version"`
}

type entitlementInput struct {
	Key       string `json:"key"`
	Kind      string `json:"kind"`
	Boolean   *bool  `json:"boolean,omitempty"`
	Integer   *int64 `json:"integer,omitempty"`
	Unlimited bool   `json:"unlimited"`
}

func (input offeringVersionInput) application() billingapi.OfferingVersionInput {
	prices := make([]billingapi.PriceInput, len(input.Prices))
	for index, price := range input.Prices {
		prices[index] = billingapi.PriceInput{
			Cadence: price.Cadence, MinorUnits: price.MinorUnits,
			Currency: price.Currency, RegistryVersion: price.RegistryVersion,
		}
	}
	entitlements := make([]billingapi.EntitlementInput, len(input.Entitlements))
	for index, value := range input.Entitlements {
		entitlements[index] = billingapi.EntitlementInput{
			Key: value.Key, Kind: value.Kind, Boolean: value.Boolean,
			Integer: value.Integer, Unlimited: value.Unlimited,
		}
	}
	return billingapi.OfferingVersionInput{
		Name: input.Name, Public: input.Public, Available: input.Available,
		Base:      input.Base,
		TrialDays: input.TrialDays, GraceDays: input.GraceDays,
		EffectiveStart: input.EffectiveStart, EffectiveEnd: input.EffectiveEnd,
		Prices: prices, Entitlements: entitlements,
	}
}

func (input offeringVersionInput) validInitial() bool {
	return requiredText(input.Name, 128) && input.TrialDays >= 0 && input.GraceDays >= 0 &&
		input.ExpectedRevision == 0 && validUTCInterval(input.EffectiveStart, input.EffectiveEnd) &&
		input.validPrices() && input.validEntitlements()
}

func (input offeringVersionInput) validMutation() bool {
	return input.ExpectedRevision > 0 && requiredText(input.Name, 128) &&
		input.TrialDays >= 0 && input.GraceDays >= 0 &&
		validUTCInterval(input.EffectiveStart, input.EffectiveEnd) &&
		input.validPrices() && input.validEntitlements()
}

func (input offeringVersionInput) validPrices() bool {
	if len(input.Prices) < 1 || len(input.Prices) > 16 {
		return false
	}
	seen := make(map[string]struct{}, len(input.Prices))
	for _, price := range input.Prices {
		if (price.Cadence != "monthly" && price.Cadence != "annual") ||
			price.MinorUnits < 0 || !validCurrency(price.Currency) || price.RegistryVersion < 1 {
			return false
		}
		if _, duplicate := seen[price.Cadence]; duplicate {
			return false
		}
		seen[price.Cadence] = struct{}{}
	}
	return true
}

func (input offeringVersionInput) validEntitlements() bool {
	if len(input.Entitlements) > 256 {
		return false
	}
	seen := make(map[string]struct{}, len(input.Entitlements))
	for _, value := range input.Entitlements {
		if !validEntitlementKey(value.Key) || !value.validValue() {
			return false
		}
		if _, duplicate := seen[value.Key]; duplicate {
			return false
		}
		seen[value.Key] = struct{}{}
	}
	return true
}

func (input entitlementInput) validValue() bool {
	if input.Kind == "boolean" {
		return input.Boolean != nil && input.Integer == nil && !input.Unlimited
	}
	if input.Kind != "integer" || input.Boolean != nil {
		return false
	}
	if input.Unlimited {
		return input.Integer == nil
	}
	return input.Integer != nil && *input.Integer >= 0
}
