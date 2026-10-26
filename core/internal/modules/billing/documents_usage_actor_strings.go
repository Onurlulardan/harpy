package billing

import (
	"context"
	"time"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *DocumentsUsage) ReserveDocumentsActor(ctx context.Context,
	scope platformtx.TenantScope, sourceOperationID string, amount int64,
	expiresAt time.Time,
) (string, bool, bool, error) {
	sourceID, err := documentsUsageStringV7(sourceOperationID)
	if err != nil {
		return "", false, false, err
	}
	result, err := service.ReserveActor(ctx, scope, sourceID, amount, expiresAt)
	if err != nil {
		return "", false, false, err
	}
	if !result.Allowed {
		return "", false, result.Replay, nil
	}
	if !documentsUsageV7(result.ID) {
		return "", false, false, usageapp.ErrSnapshot
	}
	return result.ID.String(), true, result.Replay, nil
}

func (service *DocumentsUsage) StartDocumentsActor(ctx context.Context,
	scope platformtx.TenantScope, reservationID, receiptID string,
) error {
	reservation, receipt, err := documentsUsageStringPair(reservationID, receiptID)
	if err != nil {
		return err
	}
	return service.StartActor(ctx, scope, reservation, receipt)
}

func (service *DocumentsUsage) CommitDocumentsActor(ctx context.Context,
	scope platformtx.TenantScope, reservationID, receiptID string, actualBytes int64,
) error {
	reservation, receipt, err := documentsUsageStringPair(reservationID, receiptID)
	if err != nil {
		return err
	}
	return service.CommitActor(ctx, scope, reservation, receipt, actualBytes)
}

func (service *DocumentsUsage) ReleaseDocumentsActor(ctx context.Context,
	scope platformtx.TenantScope, reservationID, receiptID string,
) error {
	reservation, receipt, err := documentsUsageStringPair(reservationID, receiptID)
	if err != nil {
		return err
	}
	return service.ReleaseActor(ctx, scope, reservation, receipt)
}
