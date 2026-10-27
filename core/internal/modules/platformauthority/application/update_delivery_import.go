package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (queue *DeliveryQueue) UpdateDeliveryImport(
	ctx context.Context,
	claim ports.DeliveryClaim,
) error {
	if queue == nil || queue.store == nil {
		return ports.ErrConflict
	}
	return queue.store.UpdateDeliveryImport(ctx, claim)
}
