package billing

import (
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

type subscriptionEventPayload struct {
	SubscriptionID    string  `json:"subscription_id"`
	ProviderAccountID string  `json:"provider_account_id"`
	OfferingVersionID string  `json:"offering_version_id"`
	PriceVersionID    string  `json:"price_version_id"`
	Currency          string  `json:"currency"`
	State             string  `json:"state"`
	PeriodStart       *string `json:"period_start"`
	PeriodEnd         *string `json:"period_end"`
	TrialEndsAt       *string `json:"trial_ends_at"`
	GraceEndsAt       *string `json:"grace_ends_at"`
	CancelAtPeriodEnd bool    `json:"cancel_at_period_end"`
	CancelAt          *string `json:"cancel_at"`
	Revision          int64   `json:"revision"`
	ChangeKind        string  `json:"change_kind"`
}

func encodeSubscriptionEvent(
	intent subscriptionapp.OutboxIntent,
	snapshot subscriptionapp.SubscriptionSnapshot,
) ([]byte, string, error) {
	payload := subscriptionEventPayload{
		SubscriptionID:    snapshot.ID.String(),
		ProviderAccountID: snapshot.ProviderAccountID.String(),
		OfferingVersionID: snapshot.OfferingVersionID.String(),
		PriceVersionID:    snapshot.PriceVersionID.String(), Currency: snapshot.Currency,
		State: string(snapshot.State), TrialEndsAt: eventInstant(snapshot.TrialEndsAt),
		GraceEndsAt:       eventInstant(snapshot.GraceEndsAt),
		CancelAtPeriodEnd: snapshot.CancelAtPeriodEnd,
		CancelAt:          eventInstant(snapshot.CancelAt), Revision: snapshot.Revision,
		ChangeKind: intent.ChangeKind,
	}
	if snapshot.Period != nil {
		start, end := snapshot.Period.Start(), snapshot.Period.End()
		payload.PeriodStart, payload.PeriodEnd = eventInstant(&start), eventInstant(&end)
	}
	return canonicalEventPayload(payload)
}

func eventInstant(value *time.Time) *string {
	if value == nil {
		return nil
	}
	encoded := value.Format(time.RFC3339Nano)
	return &encoded
}
