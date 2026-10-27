package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (queue *DeliveryQueue) RetryDelivery(
	ctx context.Context,
	claim ports.DeliveryClaim,
	errorCode string,
	permanent bool,
) error {
	if queue == nil || queue.store == nil {
		return ports.ErrConflict
	}
	return queue.store.RetryDelivery(ctx, claim, errorCode, permanent)
}
