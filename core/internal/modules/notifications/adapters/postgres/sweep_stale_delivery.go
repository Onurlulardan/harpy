package postgres

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/google/uuid"
)

func (runtime *DeliveryRuntime) SweepStaleDelivery(ctx context.Context, tenantID,
	deliveryID domain.ID, now time.Time,
) error {
	if runtime == nil || runtime.withStore == nil || ctx == nil ||
		!tenantID.Valid() || !deliveryID.Valid() || now.IsZero() {
		return ErrInvalid
	}
	tenant, err := uuid.Parse(tenantID.String())
	if err != nil {
		return ErrInvalid
	}
	return runtime.withStore(ctx, tenant,
		func(scoped context.Context, store *Store) error {
			var reconciled string
			bindErr := store.queryRow(scoped, reconcileStaleDeliverySQL,
				deliveryID.String(), now.UTC()).Scan(&reconciled)
			if store.notFound(bindErr) {
				return ErrConflict
			}
			if bindErr != nil || reconciled != deliveryID.String() {
				return ErrConflict
			}
			return nil
		})
}
