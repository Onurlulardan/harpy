package postgres

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	"github.com/google/uuid"
)

func (runtime *DeliveryRuntime) AcquireDelivery(ctx context.Context, tenantID,
	deliveryID domain.ID, owner string, now time.Time,
) (ports.DeliveryWork, error) {
	if runtime == nil || runtime.withStore == nil || ctx == nil ||
		!tenantID.Valid() || !deliveryID.Valid() || owner == "" || now.IsZero() {
		return ports.DeliveryWork{}, ErrInvalid
	}
	tenant, err := uuid.Parse(tenantID.String())
	if err != nil {
		return ports.DeliveryWork{}, ErrInvalid
	}
	var work ports.DeliveryWork
	err = runtime.withStore(ctx, tenant,
		func(scoped context.Context, store *Store) error {
			var bindErr error
			work.Delivery, bindErr = scanDelivery(store.queryRow(scoped,
				claimDeliveryWorkSQL, deliveryID.String(), owner, now.UTC()))
			if store.notFound(bindErr) {
				return ErrConflict
			}
			if bindErr != nil || work.Delivery.State == domain.DeliverySuppressed {
				return bindErr
			}
			work.Message, bindErr = scanMessage(store.queryRow(scoped,
				getWorkerMessageSQL, work.Delivery.MessageID.String()))
			if bindErr == nil {
				work.Provider, bindErr = scanProvider(store.queryRow(scoped,
					getProviderVersionSQL, work.Delivery.ProviderConfigID.String()))
			}
			if bindErr == nil {
				work.Secret, bindErr = scanProviderSecret(store.queryRow(scoped,
					getProviderSecretSQL, work.Provider.SecretID.String()))
			}
			return bindErr
		})
	return work, err
}
