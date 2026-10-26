package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) PersistMessageRevision(ctx context.Context, value domain.Message,
	expected int64,
) error {
	if !store.usable(ctx, true) || !value.ID.Valid() || !value.RecipientID.Valid() ||
		expected < 1 || value.Revision < expected {
		return ErrInvalid
	}
	result, err := store.execute(ctx, updateMessageSQL, value.ID.String(),
		value.RecipientID.String(), expected, value.Subject, value.Body, value.Revision,
		value.ReadAt, value.DismissedAt, value.PurgedAt)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrConflict
	}
	return nil
}
