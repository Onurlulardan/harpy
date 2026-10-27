package offering

import (
	"time"

	entitlementdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	moneydomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
)

type VersionInput struct {
	Name            string
	Description     string
	SortOrder       int32
	HighlightLabel  string
	Color           string
	BillingInterval string
	Public          bool
	Available       bool
	Base            bool
	TrialDays       int32
	GraceDays       int32
	EffectiveStart  time.Time
	EffectiveEnd    *time.Time
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

func ParseKey(value string) (offeringdomain.Key, error) {
	key, err := offeringdomain.NewKey(value)
	if err != nil {
		return "", ErrRequest
	}
	return key, nil
}

func BuildVersion(input VersionInput) (offeringdomain.Version, error) {
	prices, priceErr := buildInputPrices(input.Prices)
	grants, grantErr := buildInputGrants(input.Entitlements)
	if priceErr != nil || grantErr != nil {
		return offeringdomain.Version{}, ErrRequest
	}
	interval := input.BillingInterval
	if interval == "" {
		interval = string(offeringdomain.MonthlyInterval)
	}
	presentation, err := offeringdomain.NewPresentation(
		input.Description, input.SortOrder, input.HighlightLabel, input.Color,
		offeringdomain.BillingInterval(interval),
	)
	if err != nil {
		return offeringdomain.Version{}, ErrRequest
	}
	version, err := offeringdomain.NewVersionWithPresentation(
		input.Name, presentation, input.Public, input.Available, input.Base,
		input.TrialDays, input.GraceDays,
		input.EffectiveStart, input.EffectiveEnd, prices, grants,
	)
	if err != nil {
		return offeringdomain.Version{}, ErrRequest
	}
	return version, nil
}

func buildInputPrices(inputs []PriceInput) ([]offeringdomain.Price, error) {
	prices := make([]offeringdomain.Price, len(inputs))
	for index, input := range inputs {
		amount, err := moneydomain.New(input.MinorUnits, input.Currency, input.RegistryVersion)
		if err != nil {
			return nil, ErrRequest
		}
		price, err := offeringdomain.NewPrice(offeringdomain.Cadence(input.Cadence), amount)
		if err != nil {
			return nil, ErrRequest
		}
		prices[index] = price
	}
	return prices, nil
}

func buildInputGrants(inputs []EntitlementInput) ([]offeringdomain.Grant, error) {
	registry := entitlementdomain.RegistryV1ByKey()
	grants := make([]offeringdomain.Grant, len(inputs))
	for index, input := range inputs {
		key, err := entitlementdomain.NewKey(input.Key)
		entry, exists := registry[key]
		if err != nil || !exists || string(entry.Definition.ValueKind()) != input.Kind {
			return nil, ErrRequest
		}
		value, err := buildInputEntitlementValue(input)
		if err != nil {
			return nil, err
		}
		grant, err := offeringdomain.NewGrant(entry.Definition, value)
		if err != nil {
			return nil, ErrRequest
		}
		grants[index] = grant
	}
	return grants, nil
}

func buildInputEntitlementValue(input EntitlementInput) (entitlementdomain.Value, error) {
	switch input.Kind {
	case string(entitlementdomain.Boolean):
		if input.Boolean == nil || input.Integer != nil || input.Unlimited {
			return entitlementdomain.Value{}, ErrRequest
		}
		return entitlementdomain.BooleanValue(*input.Boolean), nil
	case string(entitlementdomain.Integer):
		if input.Boolean != nil || input.Unlimited && input.Integer != nil ||
			!input.Unlimited && input.Integer == nil {
			return entitlementdomain.Value{}, ErrRequest
		}
		if input.Unlimited {
			return entitlementdomain.UnlimitedValue(), nil
		}
		value, err := entitlementdomain.IntegerValue(*input.Integer)
		if err != nil {
			return entitlementdomain.Value{}, ErrRequest
		}
		return value, nil
	default:
		return entitlementdomain.Value{}, ErrRequest
	}
}
