package subscription

import (
	"encoding/json"
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

type currentOfferingWire struct {
	ID, VersionID                          uuid.UUID
	Key, State, Name, Description          string
	HighlightLabel, Color, BillingInterval string
	Revision, Version                      int64
	SortOrder                              int32
	Public, Base                           bool
	TrialDays, GraceDays                   int32
	ExpectedPriceCount, ExpectedGrantCount int
	EffectiveStart, CreatedAt, UpdatedAt   time.Time
	EffectiveEnd                           *time.Time
	Prices                                 []currentPriceWire
	Entitlements                           []currentEntitlementWire
}

type currentPriceWire struct {
	ID                      uuid.UUID
	Cadence                 string
	MinorUnits              int64
	Currency                string
	CurrencyRegistryVersion int64
}

type currentEntitlementWire struct {
	Key       string
	Kind      string
	Boolean   *bool
	Integer   *int64
	Unlimited bool
}

func decodeCurrentOffering(
	document []byte,
) (subscriptionapp.CurrentOffering, error) {
	var wire currentOfferingWire
	if len(document) == 0 || string(document) == "null" {
		return subscriptionapp.CurrentOffering{}, subscriptionapp.ErrCurrentSubscriptionInvariant
	}
	if err := json.Unmarshal(document, &wire); err != nil {
		return subscriptionapp.CurrentOffering{}, subscriptionapp.ErrCurrentSubscriptionInvariant
	}
	prices := make([]subscriptionapp.CurrentOfferingPrice, len(wire.Prices))
	for index, price := range wire.Prices {
		prices[index] = subscriptionapp.CurrentOfferingPrice{
			ID: price.ID, Cadence: price.Cadence, MinorUnits: price.MinorUnits,
			Currency:                price.Currency,
			CurrencyRegistryVersion: price.CurrencyRegistryVersion,
		}
	}
	entitlements := make([]subscriptionapp.CurrentOfferingEntitlement, len(wire.Entitlements))
	for index, entitlement := range wire.Entitlements {
		entitlements[index] = subscriptionapp.CurrentOfferingEntitlement{
			Key: entitlement.Key, Kind: entitlement.Kind,
			Boolean: entitlement.Boolean, Integer: entitlement.Integer,
			Unlimited: entitlement.Unlimited,
		}
	}
	return subscriptionapp.NewCurrentOffering(subscriptionapp.CurrentOfferingInput{
		ID: wire.ID, VersionID: wire.VersionID, Key: wire.Key, State: wire.State,
		Name: wire.Name, Description: wire.Description, Revision: wire.Revision,
		Version: wire.Version, Public: wire.Public, Base: wire.Base,
		SortOrder: wire.SortOrder, HighlightLabel: wire.HighlightLabel,
		Color: wire.Color, BillingInterval: wire.BillingInterval,
		TrialDays: wire.TrialDays, GraceDays: wire.GraceDays,
		ExpectedPriceCount: wire.ExpectedPriceCount,
		ExpectedGrantCount: wire.ExpectedGrantCount,
		EffectiveStart:     wire.EffectiveStart, EffectiveEnd: wire.EffectiveEnd,
		CreatedAt: wire.CreatedAt, UpdatedAt: wire.UpdatedAt,
		Prices: prices, Entitlements: entitlements,
	})
}
