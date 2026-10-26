package billing

import (
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

const tenantHTTPInstantLayout = "2006-01-02T15:04:05.000000Z"

func subscriptionSnapshot(
	value subscriptionapp.CurrentSubscription,
) (billingapi.TenantSubscriptionSnapshot, error) {
	period, available := value.Period()
	if !available {
		return billingapi.TenantSubscriptionSnapshot{}, billingapi.ErrUnavailable
	}
	view := billingapi.SubscriptionView{
		ID: value.ID().String(), TenantID: value.TenantID().String(),
		LegacyID: value.LegacyID().String(), LegacyPlanID: value.LegacyPlanID().String(),
		OfferingVersionID: value.OfferingVersionID().String(),
		PriceVersionID:    value.PriceVersionID().String(), Provider: value.Provider(),
		State: string(value.State()), PeriodStart: formatInstant(period.Start()),
		PeriodEnd:         formatInstant(period.End()),
		CancelAtPeriodEnd: value.EndOfPeriodCancellation(), Version: value.Revision(),
		EffectiveMinor: value.EffectiveMinor(), EffectiveCurrency: value.EffectiveCurrency(),
		EffectiveRegistry: value.EffectiveRegistry(), CreatedAt: formatInstant(value.CreatedAt()),
		UpdatedAt: formatInstant(value.UpdatedAt()),
		Offering:  currentOfferingView(value.Offering()),
	}
	if instant, found := value.TrialStart(); found {
		formatted := formatInstant(instant)
		view.TrialStartedAt = &formatted
	}
	if instant, found := value.TrialEnd(); found {
		formatted := formatInstant(instant)
		view.TrialEndsAt = &formatted
	}
	if instant, found := value.GraceEnd(); found {
		formatted := formatInstant(instant)
		view.GraceEndsAt = &formatted
	}
	if instant, found := value.Cancelled(); found {
		formatted := formatInstant(instant)
		view.CancelledAt = &formatted
	}
	view.CancellationReason = value.CancellationReason()
	return billingapi.TenantSubscriptionSnapshot{
		ID: value.ID(), TenantID: value.TenantID(),
		OfferingVersionID: value.OfferingVersionID(),
		PriceVersionID:    value.PriceVersionID(), View: view,
	}, nil
}

func formatInstant(value time.Time) string {
	return value.UTC().Format(tenantHTTPInstantLayout)
}
