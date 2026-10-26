package billing

import subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"

type entitlementEventPayload struct {
	SubscriptionID       string `json:"subscription_id"`
	SubscriptionRevision int64  `json:"subscription_revision"`
}

func encodeEntitlementEvent(
	snapshot subscriptionapp.SubscriptionSnapshot,
) ([]byte, string, error) {
	return canonicalEventPayload(entitlementEventPayload{
		SubscriptionID:       snapshot.ID.String(),
		SubscriptionRevision: snapshot.Revision,
	})
}
