package billing

import (
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

func currentOfferingView(value subscriptionapp.CurrentOffering) billingapi.OfferingView {
	prices := value.Prices()
	priceViews := make([]billingapi.PriceView, len(prices))
	for index, price := range prices {
		priceViews[index] = billingapi.PriceView{
			ID: price.ID.String(), Cadence: price.Cadence, Version: 1,
			Amount: billingapi.MoneyView{
				MinorUnits: price.MinorUnits, Currency: price.Currency,
				RegistryVersion: price.CurrencyRegistryVersion,
			},
		}
	}
	entitlements := value.Entitlements()
	entitlementViews := make([]billingapi.EntitlementValueView, len(entitlements))
	for index, entitlement := range entitlements {
		entitlementViews[index] = billingapi.EntitlementValueView{
			Key: entitlement.Key, Kind: entitlement.Kind,
			Boolean: entitlement.Boolean, Integer: entitlement.Integer,
			Unlimited: entitlement.Unlimited,
		}
	}
	return billingapi.OfferingView{
		ID: value.ID().String(), Key: value.Key(), Revision: value.Revision(),
		VersionID: value.VersionID().String(), Version: value.Version(),
		State: value.State(), Name: value.Name(), Description: value.Description(),
		SortOrder: value.SortOrder(), HighlightLabel: value.HighlightLabel(),
		Color: value.Color(), BillingInterval: value.BillingInterval(),
		Public: value.Public(), Base: value.Base(), TrialDays: int(value.TrialDays()),
		GraceDays: int(value.GraceDays()), EffectiveStart: formatInstant(value.EffectiveStart()),
		EffectiveEnd: formatOfferingEnd(value), Prices: priceViews,
		Entitlements: entitlementViews, CreatedAt: value.CreatedAt(), UpdatedAt: value.UpdatedAt(),
	}
}

func formatOfferingEnd(value subscriptionapp.CurrentOffering) *string {
	instant := value.EffectiveEnd()
	if instant == nil {
		return nil
	}
	formatted := formatInstant(*instant)
	return &formatted
}
