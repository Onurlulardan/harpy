package notifications

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type deliveryTargetResolver struct{ runtime *postgres.DeliveryRuntime }

func (resolver deliveryTargetResolver) ResolveDeliveryTarget(ctx context.Context,
	tenantID, recipientID ports.ID, channel ports.Channel,
) ([]byte, error) {
	return resolver.runtime.DeliveryTarget(ctx, tenantID, recipientID, channel)
}

var _ ports.DeliveryTargetResolver = deliveryTargetResolver{}
