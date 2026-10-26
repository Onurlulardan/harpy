package documents

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadreconcileapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/uploadreconcile"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (billing platformReconcileBilling) Release(ctx context.Context,
	scope platformtx.JobTenantScope, reservationID, receiptID foundation.ID,
) error {
	return billing.service.ReleaseDocumentsJob(ctx, scope, reservationID.String(),
		receiptID.String())
}

var _ uploadreconcileapp.Billing = platformReconcileBilling{}
