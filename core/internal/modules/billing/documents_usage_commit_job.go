package billing

import (
	"context"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *DocumentsUsage) CommitJob(
	ctx context.Context,
	scope platformtx.JobTenantScope,
	reservationID, receiptID uuid.UUID,
	actualBytes int64,
) error {
	commands, err := service.bindJobCommands(scope)
	if err != nil {
		return err
	}
	_, err = commands.Complete(ctx, usageapp.TransitionRequest{
		ReservationID: reservationID, ReceiptID: &receiptID,
		CommittedAmount: actualBytes,
	})
	return err
}
