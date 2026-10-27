package documents

import (
	"context"

	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (billing platformDeletionBilling) Decrease(ctx context.Context,
	scope platformtx.JobTenantScope, ownerFactID foundation.ID, amount int64,
	receiptID foundation.ID,
) error {
	return billing.service.DecreaseDocumentsJob(ctx, scope, ownerFactID.String(),
		amount, receiptID.String())
}

func (billing platformDeletionBilling) Release(ctx context.Context,
	scope platformtx.JobTenantScope, reservationID, receiptID foundation.ID,
) error {
	return billing.service.ReleaseDocumentsJob(ctx, scope, reservationID.String(),
		receiptID.String())
}

func (billing platformDeletionBilling) Keep(ctx context.Context,
	scope platformtx.JobTenantScope, ownerFactID, receiptID foundation.ID,
) error {
	return billing.service.KeepDocumentsJob(ctx, scope, ownerFactID.String(),
		receiptID.String())
}

var _ deletionapp.Billing = platformDeletionBilling{}
