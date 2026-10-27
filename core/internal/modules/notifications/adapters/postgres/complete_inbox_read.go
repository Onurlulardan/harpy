package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) CompleteInboxRead(ctx context.Context, recipient domain.ID,
	boundary int64, actor domain.ID,
) (int64, error) {
	if !store.usable(ctx, true) || !recipient.Valid() || !actor.Valid() || boundary < 0 {
		return 0, ErrInvalid
	}
	var count int64
	err := store.queryRow(ctx, markAllReadSQL, recipient.String(), boundary).Scan(&count)
	return count, err
}
