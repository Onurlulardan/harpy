package documents

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (billing platformActorBilling) Reserve(ctx context.Context,
	scope platformtx.TenantScope, request uploadapp.ReservationRequest,
) (uploadapp.Reservation, error) {
	id, allowed, replay, err := billing.service.ReserveDocumentsActor(ctx, scope,
		request.SourceOperationID.String(), request.Amount, request.ExpiresAt)
	if err != nil {
		return uploadapp.Reservation{}, err
	}
	if !allowed {
		return uploadapp.Reservation{Allowed: false, Replay: replay}, nil
	}
	reservationID, err := foundation.ParseID(id)
	if err != nil || !reservationID.UUIDv7() {
		return uploadapp.Reservation{}, uploadapp.ErrUnavailable
	}
	return uploadapp.Reservation{ID: reservationID,
		Allowed: true, Replay: replay}, nil
}

func (billing platformActorBilling) StartEffect(ctx context.Context,
	scope platformtx.TenantScope, reservationID, receiptID foundation.ID,
) error {
	return billing.service.StartDocumentsActor(ctx, scope, reservationID.String(),
		receiptID.String())
}

func (billing platformActorBilling) Commit(ctx context.Context,
	scope platformtx.TenantScope, reservationID, receiptID foundation.ID, actualBytes int64,
) error {
	return billing.service.CommitDocumentsActor(ctx, scope, reservationID.String(),
		receiptID.String(), actualBytes)
}

func (billing platformActorBilling) Release(ctx context.Context,
	scope platformtx.TenantScope, reservationID, receiptID foundation.ID,
) error {
	return billing.service.ReleaseDocumentsActor(ctx, scope, reservationID.String(),
		receiptID.String())
}

var _ uploadapp.Billing = platformActorBilling{}
