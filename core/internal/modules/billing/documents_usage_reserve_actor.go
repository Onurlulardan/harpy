package billing

import (
	"context"
	"time"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *DocumentsUsage) ReserveActor(
	ctx context.Context,
	scope platformtx.TenantScope,
	sourceOperationID uuid.UUID,
	amount int64,
	expiresAt time.Time,
) (DocumentsUsageReservation, error) {
	key, err := usageapp.DocumentsStorageEntitlement()
	if err != nil {
		return DocumentsUsageReservation{}, err
	}
	commands, err := service.bindActorCommands(scope)
	if err != nil {
		return DocumentsUsageReservation{}, err
	}
	result, err := commands.Reserve(ctx, usageapp.ReserveRequest{
		EntitlementKey: key, SourceOperationID: sourceOperationID,
		Amount: amount, ExpiresAt: expiresAt, ResolveCurrentPeriod: true,
	})
	if err != nil {
		return DocumentsUsageReservation{}, err
	}
	return DocumentsUsageReservation{ID: result.Reservation.ID,
		Allowed: result.Allowed, Replay: result.Replay}, nil
}
