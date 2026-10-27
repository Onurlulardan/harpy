package billing

import (
	"context"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *DocumentsUsage) StartJob(
	ctx context.Context,
	scope platformtx.JobTenantScope,
	reservationID, receiptID uuid.UUID,
) error {
	commands, err := service.bindJobCommands(scope)
	if err != nil {
		return err
	}
	_, err = commands.StartEffect(ctx, usageapp.StartEffectRequest{
		ReservationID: reservationID, ReceiptID: receiptID,
	})
	return err
}
