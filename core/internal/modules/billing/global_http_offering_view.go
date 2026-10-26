package billing

import (
	catalogquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func offeringView(snapshot catalogquery.OfferingSnapshot) billingapi.OfferingView {
	prices := snapshot.Prices()
	priceViews := make([]billingapi.PriceView, len(prices))
	for index, price := range prices {
		priceViews[index] = billingapi.PriceView{
			ID: price.ID().String(), Cadence: price.Cadence(), Version: 1,
			Amount: billingapi.MoneyView{
				MinorUnits: price.MinorUnits(), Currency: price.Currency(),
				RegistryVersion: price.CurrencyRegistryVersion(),
			},
		}
	}
	entitlements := snapshot.Entitlements()
	entitlementViews := make([]billingapi.EntitlementValueView, len(entitlements))
	for index, value := range entitlements {
		entitlementViews[index] = billingapi.EntitlementValueView{
			Key: value.Key(), Kind: value.Kind(), Boolean: value.Boolean(),
			Integer: value.Integer(), Unlimited: value.Unlimited(),
		}
	}
	return billingapi.OfferingView{
		ID: snapshot.ID().String(), Key: snapshot.Key(), Revision: snapshot.Revision(),
		VersionID: snapshot.VersionID().String(), Version: snapshot.Version(),
		State: snapshot.State(), Name: snapshot.Name(), Description: snapshot.Description(),
		SortOrder: snapshot.SortOrder(), HighlightLabel: snapshot.HighlightLabel(),
		Color: snapshot.Color(), BillingInterval: snapshot.BillingInterval(),
		Public: snapshot.Public(), Available: snapshot.Available(),
		Base: snapshot.Base(), TrialDays: int(snapshot.TrialDays()),
		GraceDays:      int(snapshot.GraceDays()),
		EffectiveStart: formatTimestamp(snapshot.EffectiveStart()),
		EffectiveEnd:   formatOptionalTimestamp(snapshot.EffectiveEnd()),
		Prices:         priceViews, Entitlements: entitlementViews,
		CreatedAt: snapshot.CreatedAt(), UpdatedAt: snapshot.UpdatedAt(),
	}
}
