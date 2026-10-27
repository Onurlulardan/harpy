package billing

import (
	"context"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *DocumentsUsage) CommitActor(
	ctx context.Context,
	scope platformtx.TenantScope,
	reservationID, receiptID uuid.UUID,
	actualBytes int64,
) error {
	commands, err := service.bindActorCommands(scope)
	if err != nil {
		return err
	}
	_, err = commands.Complete(ctx, usageapp.TransitionRequest{
		ReservationID: reservationID, ReceiptID: &receiptID,
		CommittedAmount: actualBytes,
	})
	return err
}
