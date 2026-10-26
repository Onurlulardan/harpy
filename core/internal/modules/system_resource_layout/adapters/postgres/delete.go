package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (store *layoutStore) Delete(
	ctx context.Context,
	id uuid.UUID,
	expectedRowRevision int64,
) error {
	if !store.commandUsable(ctx) {
		return layoutUnavailable()
	}
	var deleted bool
	if err := store.queryRow(
		ctx, deleteSQL, id, expectedRowRevision,
	).Scan(&deleted); err != nil {
		return mapLayoutError(err)
	}
	if !deleted {
		return layoutUnavailable()
	}
	return nil
}
