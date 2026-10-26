package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/google/uuid"
)

func (storage *store) writeEffectStart(
	ctx context.Context,
	change usageapp.EffectStart,
) error {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return err
	}
	var reservationID uuid.UUID
	err = storage.queryRow(ctx, startEffectSQL, authorityID, storage.tenantID,
		change.ReservationID, change.ExpectedRevision, change.StartedAt, change.ReceiptID,
	).Scan(&reservationID)
	if err != nil {
		return fmt.Errorf("update billing usage effect start: %w", err)
	}
	if reservationID != change.ReservationID {
		return usageapp.ErrSnapshot
	}
	return nil
}
