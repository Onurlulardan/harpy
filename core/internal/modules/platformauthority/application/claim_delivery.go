package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (queue *DeliveryQueue) ClaimDelivery(
	ctx context.Context,
	ownerID string,
) (ports.DeliveryClaim, bool, error) {
	if queue == nil || queue.store == nil {
		return ports.DeliveryClaim{}, false, ports.ErrConflict
	}
	return queue.store.ClaimDelivery(ctx, ownerID)
}
