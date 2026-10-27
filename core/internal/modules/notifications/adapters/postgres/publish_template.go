package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) PublishTemplate(ctx context.Context, templateID,
	versionID domain.ID, expected int64,
) error {
	if !store.usable(ctx, true) || !templateID.Valid() || !versionID.Valid() ||
		expected < 1 {
		return ErrInvalid
	}
	result, err := store.execute(ctx, publishTemplateSQL, templateID.String(),
		versionID.String(), expected)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrConflict
	}
	return nil
}
