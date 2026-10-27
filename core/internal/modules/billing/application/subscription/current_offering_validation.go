package subscriptionapp

import (
	"bytes"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var currentOfferingKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,126}[a-z0-9]$`)
var currentEntitlementKeyPattern = regexp.MustCompile(`^[a-z](?:[a-z0-9_.]{0,126}[a-z0-9])?$`)

func (value CurrentOffering) valid(priceCount, grantCount int) bool {
	if !validID(value.id) || !validID(value.versionID) ||
		!currentOfferingKeyPattern.MatchString(value.key) ||
		(value.state != "active" && value.state != "retired") ||
		value.revision < 1 || value.version < 1 || !currentOfferingName(value.name) ||
		!currentOfferingText(value.description, 2048, false) ||
		!currentOfferingText(value.highlightLabel, 50, true) ||
		!currentOfferingText(value.color, 20, true) ||
		(value.billingInterval != "monthly" && value.billingInterval != "annual") ||
		value.trialDays < 0 || value.trialDays > 3650 ||
		value.graceDays < 0 || value.graceDays > 3650 ||
		!validInstant(value.effectiveStart) || !validInstant(value.createdAt) ||
		!validInstant(value.updatedAt) || value.updatedAt.Before(value.createdAt) ||
		value.effectiveEnd != nil && (!validInstant(*value.effectiveEnd) ||
			!value.effectiveStart.Before(*value.effectiveEnd)) ||
		priceCount != len(value.prices) || grantCount != len(value.entitlements) ||
		priceCount < 1 || priceCount > 16 || grantCount < 0 || grantCount > 256 {
		return false
	}
	return validCurrentPrices(value.prices) && validCurrentEntitlements(value.entitlements)
}

func validCurrentPrices(values []CurrentOfferingPrice) bool {
	for index, value := range values {
		if !validID(value.ID) ||
			(value.Cadence != "one_time" && value.Cadence != "monthly" &&
				value.Cadence != "yearly") ||
			value.MinorUnits < 0 || value.MinorUnits > 9000000000000000 ||
			!currencyPattern.MatchString(value.Currency) ||
			value.CurrencyRegistryVersion < 1 ||
			index > 0 && compareCurrentPrice(values[index-1], value) >= 0 {
			return false
		}
	}
	return true
}

func validCurrentEntitlements(values []CurrentOfferingEntitlement) bool {
	for index, value := range values {
		validValue := value.Kind == "boolean" && value.Boolean != nil &&
			value.Integer == nil && !value.Unlimited ||
			value.Kind == "integer" && value.Boolean == nil &&
				((value.Unlimited && value.Integer == nil) ||
					(!value.Unlimited && value.Integer != nil && *value.Integer >= 0))
		if !currentEntitlementKeyPattern.MatchString(value.Key) || !validValue ||
			index > 0 && values[index-1].Key >= value.Key {
			return false
		}
	}
	return true
}

func compareCurrentPrice(left, right CurrentOfferingPrice) int {
	if value := strings.Compare(left.Cadence, right.Cadence); value != 0 {
		return value
	}
	if value := strings.Compare(left.Currency, right.Currency); value != 0 {
		return value
	}
	return bytes.Compare(left.ID[:], right.ID[:])
}

func currentOfferingName(value string) bool {
	return value != "" && len(value) <= 128 && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value
}

func currentOfferingText(value string, maximum int, trim bool) bool {
	return utf8.ValidString(value) && len(value) <= maximum &&
		(!trim || strings.TrimSpace(value) == value)
}

func copyOfferingPrices(values []CurrentOfferingPrice) []CurrentOfferingPrice {
	return append([]CurrentOfferingPrice(nil), values...)
}

func copyOfferingEntitlements(
	values []CurrentOfferingEntitlement,
) []CurrentOfferingEntitlement {
	result := make([]CurrentOfferingEntitlement, len(values))
	for index, value := range values {
		result[index] = value
		result[index].Boolean = copyBool(value.Boolean)
		result[index].Integer = copyInt64(value.Integer)
	}
	return result
}

func copyBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func copyInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func (value CurrentOffering) hasPrice(identifier uuid.UUID) bool {
	for _, price := range value.prices {
		if price.ID == identifier {
			return true
		}
	}
	return false
}
