package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	"github.com/google/uuid"
)

func (runtime *DeliveryRuntime) PersistProviderReceipt(ctx context.Context,
	tenantID domain.ID, receipt ports.AuthenticatedProviderReceipt,
) (replay bool, err error) {
	if runtime == nil || runtime.withStore == nil || ctx == nil ||
		!tenantID.Valid() || !validProviderReceipt(receipt) {
		return false, ErrInvalid
	}
	tenant, parseErr := uuid.Parse(tenantID.String())
	if parseErr != nil {
		return false, ErrInvalid
	}
	err = runtime.withStore(ctx, tenant,
		func(scoped context.Context, store *Store) error {
			var bindErr error
			replay, bindErr = store.PersistProviderReceipt(scoped, receipt)
			return bindErr
		})
	return replay, err
}
