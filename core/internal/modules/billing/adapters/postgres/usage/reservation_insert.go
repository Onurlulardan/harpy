package usage

import (
	"context"
	"fmt"
	"time"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/google/uuid"
)

func (storage *store) writeReservation(
	ctx context.Context,
	reservation usageapp.Reservation,
	createdAt time.Time,
) error {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return err
	}
	var insertedID uuid.UUID
	err = storage.queryRow(ctx, insertReservationSQL, authorityID, storage.tenantID,
		reservation.ID, reservation.DecisionID, string(reservation.EntitlementKey),
		reservation.PeriodID, reservation.SourceOperationID, reservation.Amount,
		reservation.ExpiresAt, createdAt,
	).Scan(&insertedID)
	if err != nil {
		return fmt.Errorf("insert billing usage reservation: %w", err)
	}
	if insertedID != reservation.ID {
		return usageapp.ErrSnapshot
	}
	return nil
}
