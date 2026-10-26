package billing

import (
	"context"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *DocumentsUsage) ReleaseActor(
	ctx context.Context,
	scope platformtx.TenantScope,
	reservationID, receiptID uuid.UUID,
) error {
	commands, err := service.bindActorCommands(scope)
	if err != nil {
		return err
	}
	_, err = commands.Release(ctx, usageapp.TransitionRequest{
		ReservationID: reservationID, ReceiptID: &receiptID,
	})
	return err
}
