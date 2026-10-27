package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (store *Store) RetryDelivery(
	ctx context.Context,
	claim ports.DeliveryClaim,
	errorCode string,
	permanent bool,
) error {
	return store.deliveryDisposition(
		ctx, retryDeliverySQL, claim, errorCode, permanent,
	)
}
