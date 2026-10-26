package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func persistIntent(ctx context.Context, stores ports.Stores, plan intentPlan,
	actor domain.ID,
) error {
	if err := stores.Messages().RecordMessage(ctx, plan.message); err != nil {
		return ErrUnavailable
	}
	if err := stores.Effects().InsertEffect(ctx, plan.receipt); err != nil {
		return ErrUnavailable
	}
	for _, planned := range plan.deliveries {
		if err := stores.Deliveries().InsertDelivery(ctx, planned.delivery); err != nil {
			return ErrUnavailable
		}
		if planned.job != nil {
			if err := stores.Evidence().EnqueueDelivery(ctx, *planned.job); err != nil {
				return ErrUnavailable
			}
		}
		if planned.alert {
			if err := stores.Evidence().AppendOperatorAlert(ctx,
				planned.delivery.OutcomeCode, planned.delivery.ID); err != nil {
				return ErrUnavailable
			}
		}
	}
	if plan.hint != nil {
		if err := stores.Realtime().AppendHint(ctx, *plan.hint); err != nil {
			return ErrUnavailable
		}
	}
	if err := stores.Evidence().AppendAudit(ctx, "notifications.intent.create",
		plan.message.ID, actor, plan.message.Revision); err != nil {
		return ErrUnavailable
	}
	return nil
}
