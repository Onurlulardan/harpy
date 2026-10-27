package documents

import (
	"context"
	"time"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (billing platformArchiveBilling) Reserve(ctx context.Context,
	scope platformtx.JobTenantScope, reservationID, sourceID foundation.ID,
	amount int64, expiresAt time.Time,
) error {
	allowed, err := billing.service.ReserveDocumentsJob(ctx, scope,
		reservationID.String(), sourceID.String(), amount, expiresAt)
	if err == nil && !allowed {
		return archiveapp.ErrLimit
	}
	return err
}

func (billing platformArchiveBilling) StartEffect(ctx context.Context,
	scope platformtx.JobTenantScope, reservationID, receiptID foundation.ID,
) error {
	return billing.service.StartDocumentsJob(ctx, scope, reservationID.String(),
		receiptID.String())
}

func (billing platformArchiveBilling) Commit(ctx context.Context,
	scope platformtx.JobTenantScope, reservationID, receiptID foundation.ID, actualBytes int64,
) error {
	return billing.service.CommitDocumentsJob(ctx, scope, reservationID.String(),
		receiptID.String(), actualBytes)
}

var _ archiveapp.Billing = platformArchiveBilling{}
