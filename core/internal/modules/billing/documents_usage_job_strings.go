package billing

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *DocumentsUsage) ReserveDocumentsJob(ctx context.Context,
	scope platformtx.JobTenantScope, reservationID, sourceOperationID string,
	amount int64, expiresAt time.Time,
) (bool, error) {
	reservation, source, err := documentsUsageStringPair(
		reservationID, sourceOperationID,
	)
	if err != nil {
		return false, err
	}
	err = service.ReserveJob(ctx, scope, reservation, source, amount, expiresAt)
	return documentsUsageAllowed(err)
}

func (service *DocumentsUsage) StartDocumentsJob(ctx context.Context,
	scope platformtx.JobTenantScope, reservationID, receiptID string,
) error {
	reservation, receipt, err := documentsUsageStringPair(reservationID, receiptID)
	if err != nil {
		return err
	}
	return service.StartJob(ctx, scope, reservation, receipt)
}

func (service *DocumentsUsage) CommitDocumentsJob(ctx context.Context,
	scope platformtx.JobTenantScope, reservationID, receiptID string, actualBytes int64,
) error {
	reservation, receipt, err := documentsUsageStringPair(reservationID, receiptID)
	if err != nil {
		return err
	}
	return service.CommitJob(ctx, scope, reservation, receipt, actualBytes)
}

func (service *DocumentsUsage) ReleaseDocumentsJob(ctx context.Context,
	scope platformtx.JobTenantScope, reservationID, receiptID string,
) error {
	reservation, receipt, err := documentsUsageStringPair(reservationID, receiptID)
	if err != nil {
		return err
	}
	return service.ReleaseJob(ctx, scope, reservation, receipt)
}

func (service *DocumentsUsage) DecreaseDocumentsJob(ctx context.Context,
	scope platformtx.JobTenantScope, ownerFactID string, amount int64, receiptID string,
) error {
	owner, receipt, err := documentsUsageStringPair(ownerFactID, receiptID)
	if err != nil {
		return err
	}
	return service.DecreaseJob(ctx, scope, owner, amount, receipt)
}

func (service *DocumentsUsage) KeepDocumentsJob(ctx context.Context,
	scope platformtx.JobTenantScope, ownerFactID, receiptID string,
) error {
	owner, receipt, err := documentsUsageStringPair(ownerFactID, receiptID)
	if err != nil {
		return err
	}
	return service.KeepJob(ctx, scope, owner, receipt)
}

func documentsUsageStringPair(left, right string) (uuid.UUID, uuid.UUID, error) {
	first, err := documentsUsageStringV7(left)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	second, err := documentsUsageStringV7(right)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return first, second, nil
}
