package billing

import (
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

func platformSubscriptionView(
	snapshot subscriptionapp.SubscriptionSnapshot,
	provider string,
) billingapi.SubscriptionView {
	result := billingapi.SubscriptionView{
		ID: snapshot.ID.String(), TenantID: snapshot.TenantID.String(),
		OfferingVersionID: snapshot.OfferingVersionID.String(),
		PriceVersionID:    snapshot.PriceVersionID.String(), Provider: provider,
		State: string(snapshot.State), CancelAtPeriodEnd: snapshot.CancelAtPeriodEnd,
		Version: snapshot.Revision,
	}
	if snapshot.Period != nil {
		result.PeriodStart = formatInstant(snapshot.Period.Start())
		result.PeriodEnd = formatInstant(snapshot.Period.End())
	}
	if snapshot.TrialEndsAt != nil {
		value := formatInstant(*snapshot.TrialEndsAt)
		result.TrialEndsAt = &value
	}
	if snapshot.GraceEndsAt != nil {
		value := formatInstant(*snapshot.GraceEndsAt)
		result.GraceEndsAt = &value
	}
	return result
}
