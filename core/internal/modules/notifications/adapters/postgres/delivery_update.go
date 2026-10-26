package postgres

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) updateDelivery(ctx context.Context, value domain.Delivery,
	expected int64,
) error {
	if !store.usable(ctx, true) || !value.ID.Valid() || expected < 1 ||
		value.Revision <= expected {
		return ErrInvalid
	}
	result, err := store.execute(ctx, updateDeliverySQL, value.ID.String(), expected,
		string(value.State), value.Attempt, value.LeaseOwner, value.LeaseGeneration,
		nullableTime(value.LeaseUntil), value.NotBefore, value.OutcomeCode,
		value.OutcomeUnknown, value.Revision)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrConflict
	}
	return nil
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC()
}
