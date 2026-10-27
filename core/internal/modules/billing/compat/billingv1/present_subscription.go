package billingv1

import (
	"context"
	"encoding/json"
	"fmt"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) presentSubscription(
	ctx context.Context,
	view billingapi.SubscriptionView,
) (subscriptionWire, error) {
	identity, err := handler.resolvePublicPlanIdentity(ctx, view.Offering.ID)
	if err != nil {
		return subscriptionWire{}, err
	}
	if identity.CanonicalID != view.Offering.ID {
		return subscriptionWire{}, ErrSnapshotRequired
	}
	view.LegacyPlanID = identity.LegacyID
	return legacySubscription(view, view.Offering)
}

func materializeSubscription(
	view billingapi.SubscriptionView,
) (billingapi.SubscriptionView, error) {
	document, err := json.Marshal(view)
	if err != nil {
		return billingapi.SubscriptionView{}, fmt.Errorf(
			"marshal Billing subscription query result: %w", err,
		)
	}
	var result billingapi.SubscriptionView
	if err := json.Unmarshal(document, &result); err != nil {
		return billingapi.SubscriptionView{}, fmt.Errorf(
			"decode Billing subscription query result: %w", err,
		)
	}
	return result, nil
}
