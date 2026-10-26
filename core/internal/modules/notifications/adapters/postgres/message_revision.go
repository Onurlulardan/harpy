package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) NextInboxRevision(ctx context.Context,
	recipient domain.ID,
) (int64, error) {
	if !store.usable(ctx, true) || !recipient.Valid() {
		return 0, ErrInvalid
	}
	var revision int64
	err := store.queryRow(ctx, nextInboxRevisionSQL, recipient.String()).Scan(&revision)
	return revision, err
}

func (store *Store) CurrentInboxRevision(ctx context.Context,
	recipient domain.ID,
) (int64, error) {
	if !store.usable(ctx, false) || !recipient.Valid() {
		return 0, ErrInvalid
	}
	var revision int64
	err := store.queryRow(ctx, currentInboxRevisionSQL, recipient.String()).Scan(&revision)
	if store.notFound(err) {
		return 0, nil
	}
	return revision, err
}
