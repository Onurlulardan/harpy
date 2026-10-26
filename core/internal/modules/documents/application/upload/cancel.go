package upload

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Cancel(ctx context.Context, scope platformtx.TenantScope,
	request CancelRequest,
) (Snapshot, error) {
	if !request.UploadID.UUIDv7() {
		return Snapshot{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, Command, OperationCancel, PermissionCancel)
	if err != nil {
		return Snapshot{}, err
	}
	return application.cancelBound(ctx, scope, store, request.UploadID)
}

func (application *Application) cancelBound(ctx context.Context, scope platformtx.TenantScope,
	store Store, uploadID domain.ID,
) (Snapshot, error) {
	ids, err := application.nextIDs(1)
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now()
	result, err := store.Cancel(ctx, uploadID, ids[0], now)
	if err != nil {
		return Snapshot{}, mapStoreError(err)
	}
	if result.ReleaseReservation {
		if result.BillingReservationID == (domain.ID{}) {
			return Snapshot{}, ErrUnavailable
		}
		if err = application.billing.Release(ctx, scope,
			result.BillingReservationID, ids[0]); err != nil {
			return Snapshot{}, mapBillingError(err)
		}
	}
	return result.Snapshot, nil
}
