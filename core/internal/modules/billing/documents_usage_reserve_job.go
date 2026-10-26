package billing

import (
	"context"
	"time"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *DocumentsUsage) ReserveJob(
	ctx context.Context,
	scope platformtx.JobTenantScope,
	reservationID, sourceOperationID uuid.UUID,
	amount int64,
	expiresAt time.Time,
) error {
	key, err := usageapp.DocumentsStorageEntitlement()
	if err != nil {
		return err
	}
	commands, err := service.bindJobCommands(scope)
	if err != nil {
		return err
	}
	result, err := commands.Reserve(ctx, usageapp.ReserveRequest{
		EntitlementKey: key, SourceOperationID: sourceOperationID,
		ReservationID: reservationID, Amount: amount, ExpiresAt: expiresAt,
		ResolveCurrentPeriod: true,
	})
	if err != nil {
		return err
	}
	if !result.Allowed {
		return ErrDocumentsUsageDenied
	}
	if result.Reservation.ID != reservationID {
		return usageapp.ErrSnapshot
	}
	return nil
}
