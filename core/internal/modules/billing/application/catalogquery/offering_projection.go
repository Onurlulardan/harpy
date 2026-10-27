package catalogquery

import (
	"bytes"
	"strings"
	"unicode/utf8"

	entitlementdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
)

func projectOffering(record OfferingRecord, instantValidAtRecord bool) (OfferingSnapshot, bool) {
	key, err := offeringdomain.NewKey(record.Key)
	presentation, presentationErr := offeringdomain.NewPresentation(
		record.Description, record.SortOrder, record.HighlightLabel, record.Color,
		offeringdomain.BillingInterval(record.BillingInterval),
	)
	if err != nil || record.ID == [16]byte{} || record.VersionID == [16]byte{} ||
		presentationErr != nil || !presentation.Valid() ||
		record.Revision < 1 || record.Version < 1 || !validName(record.Name) ||
		record.TrialDays < 0 || record.TrialDays > 3650 ||
		record.GraceDays < 0 || record.GraceDays > 3650 ||
		record.ExpectedPriceCount != len(record.Prices) ||
		record.ExpectedGrantCount != len(record.Entitlements) ||
		record.ExpectedPriceCount < 1 || record.ExpectedPriceCount > 16 ||
		record.ExpectedGrantCount < 0 || record.ExpectedGrantCount > 256 ||
		!validInstant(record.EffectiveStart) || !validInterval(record.EffectiveStart, record.EffectiveEnd) ||
		!validInstant(record.CreatedAt) || !validInstant(record.UpdatedAt) ||
		record.UpdatedAt.Before(record.CreatedAt) ||
		!instantValidAtRecord {
		return OfferingSnapshot{}, false
	}
	prices, valid := projectPrices(record.Prices)
	if !valid {
		return OfferingSnapshot{}, false
	}
	entitlements, valid := projectEntitlements(record.Entitlements)
	if !valid {
		return OfferingSnapshot{}, false
	}
	return OfferingSnapshot{id: record.ID, key: string(key), revision: record.Revision,
		versionID: record.VersionID, version: record.Version, name: record.Name,
		description: record.Description, sortOrder: record.SortOrder,
		highlightLabel: record.HighlightLabel, color: record.Color,
		billingInterval: record.BillingInterval, state: record.State,
		public: record.Public, available: record.Available, base: record.Base,
		trialDays: record.TrialDays, graceDays: record.GraceDays,
		effectiveStart: record.EffectiveStart, effectiveEnd: copyTime(record.EffectiveEnd),
		createdAt: record.CreatedAt, updatedAt: record.UpdatedAt,
		prices: prices, entitlements: entitlements}, true
}

func projectPrices(records []PriceRecord) ([]PriceSnapshot, bool) {
	result := make([]PriceSnapshot, len(records))
	for index, record := range records {
		if record.ID == [16]byte{} || !validCadence(record.Cadence) ||
			record.MinorUnits < 0 || !validCurrency(record.Currency) ||
			record.CurrencyRegistryVersion < 1 ||
			(index > 0 && comparePrice(records[index-1], record) >= 0) {
			return nil, false
		}
		result[index] = PriceSnapshot{id: record.ID, cadence: record.Cadence,
			minorUnits: record.MinorUnits, currency: record.Currency,
			currencyRegistryVersion: record.CurrencyRegistryVersion}
	}
	return result, true
}

func projectEntitlements(records []EntitlementRecord) ([]EntitlementSnapshot, bool) {
	result := make([]EntitlementSnapshot, len(records))
	for index, record := range records {
		key, err := entitlementdomain.NewKey(record.Key)
		if err != nil || !validEntitlementValue(record) ||
			(index > 0 && records[index-1].Key >= record.Key) {
			return nil, false
		}
		result[index] = EntitlementSnapshot{key: string(key), kind: record.Kind,
			boolean: copyBool(record.Boolean), integer: copyInt64(record.Integer),
			unlimited: record.Unlimited}
	}
	return result, true
}

func validEntitlementValue(record EntitlementRecord) bool {
	return record.Kind == "boolean" && record.Boolean != nil && record.Integer == nil &&
		!record.Unlimited || record.Kind == "integer" && record.Boolean == nil &&
		((record.Unlimited && record.Integer == nil) ||
			(!record.Unlimited && record.Integer != nil && *record.Integer >= 0))
}

func comparePrice(left, right PriceRecord) int {
	if value := strings.Compare(left.Cadence, right.Cadence); value != 0 {
		return value
	}
	if value := strings.Compare(left.Currency, right.Currency); value != 0 {
		return value
	}
	return bytes.Compare(left.ID[:], right.ID[:])
}

func validName(value string) bool {
	return value != "" && len(value) <= 128 && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value
}

func validCadence(value string) bool {
	return value == "one_time" || value == "monthly" || value == "yearly"
}

func validCurrency(value string) bool {
	return len(value) == 3 && value[0] >= 'A' && value[0] <= 'Z' &&
		value[1] >= 'A' && value[1] <= 'Z' && value[2] >= 'A' && value[2] <= 'Z'
}
