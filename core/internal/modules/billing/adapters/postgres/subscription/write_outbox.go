package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

func (storage *store) writeOutbox(
	ctx context.Context,
	plan subscriptionapp.CommitPlan,
) error {
	for _, event := range plan.Outbox {
		if err := storage.effects.PublishSubscription(ctx, event, plan.After); err != nil {
			return fmt.Errorf("publish billing subscription event: %w", err)
		}
	}
	return nil
}
