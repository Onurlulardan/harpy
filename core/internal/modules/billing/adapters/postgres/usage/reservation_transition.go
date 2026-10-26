package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/google/uuid"
)

func (storage *store) writeTransition(
	ctx context.Context,
	change usageapp.ReservationTransition,
) error {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return err
	}
	var reservationID uuid.UUID
	err = storage.queryRow(ctx, transitionReservationSQL, authorityID, storage.tenantID,
		change.LedgerEntryID, change.ReservationID, change.ExpectedRevision,
		string(change.From), string(change.To), change.ResolvedAt, change.ReceiptID,
		change.CommittedAmount, change.RemainderEntryID,
	).Scan(&reservationID)
	if err != nil {
		return fmt.Errorf("transition billing usage reservation: %w", err)
	}
	if reservationID != change.ReservationID {
		return usageapp.ErrSnapshot
	}
	return nil
}
