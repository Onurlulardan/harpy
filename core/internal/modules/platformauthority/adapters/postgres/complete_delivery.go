package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (store *Store) CompleteDelivery(
	ctx context.Context,
	claim ports.DeliveryClaim,
) error {
	return store.deliveryDisposition(ctx, succeedDeliverySQL, claim)
}
