package billing

import (
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
)

func globalHTTPOfferingInput(
	input billingapi.OfferingVersionInput,
) (offeringapp.VersionInput, error) {
	start, err := parseTimestamp(input.EffectiveStart)
	end, endErr := parseOptionalTimestamp(input.EffectiveEnd)
	trialDays, trialErr := int32Value(input.TrialDays)
	graceDays, graceErr := int32Value(input.GraceDays)
	if err != nil || endErr != nil || trialErr != nil || graceErr != nil {
		return offeringapp.VersionInput{}, errInvalidInput
	}
	prices := make([]offeringapp.PriceInput, len(input.Prices))
	for index, price := range input.Prices {
		prices[index] = offeringapp.PriceInput{
			Cadence: price.Cadence, MinorUnits: price.MinorUnits,
			Currency: price.Currency, RegistryVersion: price.RegistryVersion,
		}
	}
	entitlements := make([]offeringapp.EntitlementInput, len(input.Entitlements))
	for index, entitlement := range input.Entitlements {
		entitlements[index] = offeringapp.EntitlementInput{
			Key: entitlement.Key, Kind: entitlement.Kind, Boolean: entitlement.Boolean,
			Integer: entitlement.Integer, Unlimited: entitlement.Unlimited,
		}
	}
	return offeringapp.VersionInput{
		Name: input.Name, Description: input.Description, SortOrder: input.SortOrder,
		HighlightLabel: input.HighlightLabel, Color: input.Color,
		BillingInterval: input.BillingInterval, Public: input.Public,
		Available: input.Available, Base: input.Base,
		TrialDays: trialDays, GraceDays: graceDays, EffectiveStart: start,
		EffectiveEnd: end, Prices: prices, Entitlements: entitlements,
	}, nil
}
