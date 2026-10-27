package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	"github.com/google/uuid"
)

func (runtime *DeliveryRuntime) CommitDelivery(ctx context.Context, tenantID domain.ID,
	final ports.DeliveryFinalization, expected int64,
) error {
	if runtime == nil || runtime.withStore == nil || ctx == nil ||
		!tenantID.Valid() || !final.Delivery.ID.Valid() || expected < 1 ||
		final.Attempt.DeliveryID != final.Delivery.ID {
		return ErrInvalid
	}
	tenant, err := uuid.Parse(tenantID.String())
	if err != nil {
		return ErrInvalid
	}
	return runtime.withStore(ctx, tenant,
		func(scoped context.Context, store *Store) error {
			var bindErr error
			if bindErr = store.updateDelivery(scoped, final.Delivery, expected); bindErr != nil {
				return bindErr
			}
			if bindErr = store.InsertAttempt(scoped, final.Attempt); bindErr != nil {
				return bindErr
			}
			if final.Receipt != nil {
				if bindErr = store.InsertReceipt(scoped, *final.Receipt); bindErr != nil {
					return bindErr
				}
			}
			if final.ProviderLocator != nil {
				return store.insertProviderLocator(scoped, tenant, *final.ProviderLocator)
			}
			return nil
		})
}
