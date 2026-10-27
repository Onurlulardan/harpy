package billingv1

import (
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
)

func mapPlanCreate(
	access billingapi.Access,
	idempotencyKey string,
	input planRequest,
	now time.Time,
) (billingapi.CreateOfferingCommand, error) {
	if input.Name == nil || input.Slug == nil || !required(*input.Name) ||
		!required(*input.Slug) {
		return billingapi.CreateOfferingCommand{}, ErrRequestInvalid
	}
	limits, features, err := decodePlanDocuments(input.Limits, input.Features)
	if err != nil {
		return billingapi.CreateOfferingCommand{}, err
	}
	monthly, annual, trial := int64Value(input.PriceCents),
		int64Value(input.AnnualPriceCents), intValue(input.TrialDays)
	sortOrder := intValue(input.SortOrder)
	currency, interval := stringValue(input.Currency, "TRY"),
		stringValue(input.BillingInterval, "monthly")
	if monthly < 0 || annual < 0 || trial < 0 ||
		(interval != "monthly" && interval != "annual") ||
		interval == "annual" && annual <= 0 ||
		int64(sortOrder) < -1<<31 || int64(sortOrder) > 1<<31-1 {
		return billingapi.CreateOfferingCommand{}, ErrRequestInvalid
	}
	version := billingapi.OfferingVersionInput{
		Name: *input.Name, Description: stringValue(input.Description, ""),
		SortOrder: int32(sortOrder), HighlightLabel: stringValue(input.HighlightLabel, ""),
		Color: stringValue(input.Color, ""), BillingInterval: interval,
		Public:    boolValue(input.IsPublic, true),
		Available: boolValue(input.IsActive, true), TrialDays: trial,
		EffectiveStart: utcMicrosecond(now),
		Entitlements:   legacyEntitlements(limits, features),
		Prices: []billingapi.PriceInput{{
			Cadence: "monthly", MinorUnits: monthly, Currency: currency,
			RegistryVersion: money.RegistryVersion,
		}},
	}
	if annual > 0 {
		version.Prices = append(version.Prices, billingapi.PriceInput{
			Cadence: "yearly", MinorUnits: annual, Currency: currency,
			RegistryVersion: money.RegistryVersion,
		})
	}
	return billingapi.CreateOfferingCommand{
		Access: access, IdempotencyKey: idempotencyKey,
		Key: *input.Slug, InitialVersion: version,
	}, nil
}

func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func stringValue(value *string, fallback string) string {
	if value == nil || *value == "" {
		return fallback
	}
	return *value
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
