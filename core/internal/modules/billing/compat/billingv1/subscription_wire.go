package billingv1

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"

type subscriptionWire struct {
	ID                  string   `json:"id"`
	TenantID            string   `json:"tenant_id"`
	PlanID              string   `json:"plan_id"`
	Plan                planWire `json:"plan"`
	Status              string   `json:"status"`
	CurrentPeriodStart  string   `json:"current_period_start"`
	CurrentPeriodEnd    string   `json:"current_period_end"`
	TrialStart          *string  `json:"trial_start,omitempty"`
	TrialEnd            *string  `json:"trial_end,omitempty"`
	CancelAtPeriodEnd   bool     `json:"cancel_at_period_end"`
	CancelledAt         *string  `json:"cancelled_at,omitempty"`
	CancellationReason  string   `json:"cancellation_reason,omitempty"`
	EffectivePriceCents int64    `json:"effective_price_cents"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
}

func legacySubscription(
	view billingapi.SubscriptionView,
	offering billingapi.OfferingView,
) (subscriptionWire, error) {
	price, err := priceByID(offering, view.PriceVersionID)
	if err != nil {
		return subscriptionWire{}, err
	}
	if view.OfferingVersionID != offering.VersionID ||
		view.EffectiveMinor < 0 || view.EffectiveCurrency != price.Amount.Currency ||
		view.EffectiveRegistry != price.Amount.RegistryVersion ||
		!canonicalWireInstant(view.CreatedAt) || !canonicalWireInstant(view.UpdatedAt) {
		return subscriptionWire{}, ErrSnapshotRequired
	}
	periodStart, err := v1WireInstant(view.PeriodStart)
	if err != nil {
		return subscriptionWire{}, ErrSnapshotRequired
	}
	periodEnd, err := v1WireInstant(view.PeriodEnd)
	if err != nil {
		return subscriptionWire{}, ErrSnapshotRequired
	}
	createdAt, _ := v1WireInstant(view.CreatedAt)
	updatedAt, _ := v1WireInstant(view.UpdatedAt)
	trialStart, err := optionalV1WireInstant(view.TrialStartedAt)
	if err != nil {
		return subscriptionWire{}, ErrSnapshotRequired
	}
	trialEnd, err := optionalV1WireInstant(view.TrialEndsAt)
	if err != nil {
		return subscriptionWire{}, ErrSnapshotRequired
	}
	cancelledAt, err := optionalV1WireInstant(view.CancelledAt)
	if err != nil {
		return subscriptionWire{}, ErrSnapshotRequired
	}
	legacyID, legacyPlanID := view.LegacyID, view.LegacyPlanID
	if legacyID == "" {
		legacyID = view.ID
	}
	if legacyPlanID == "" {
		legacyPlanID = offering.ID
	}
	legacyOffering := offering
	legacyOffering.ID = legacyPlanID
	return subscriptionWire{
		ID: legacyID, TenantID: view.TenantID, PlanID: legacyPlanID,
		Plan: legacyPlan(legacyOffering), Status: view.State,
		CurrentPeriodStart: periodStart, CurrentPeriodEnd: periodEnd,
		TrialStart: trialStart, TrialEnd: trialEnd,
		CancelAtPeriodEnd: view.CancelAtPeriodEnd,
		CancelledAt:       cancelledAt, CancellationReason: view.CancellationReason,
		EffectivePriceCents: view.EffectiveMinor,
		CreatedAt:           createdAt, UpdatedAt: updatedAt,
	}, nil
}

func priceByID(
	offering billingapi.OfferingView,
	id string,
) (billingapi.PriceView, error) {
	for _, price := range offering.Prices {
		if price.ID == id {
			return price, nil
		}
	}
	return billingapi.PriceView{}, ErrPlanChangeTarget
}
