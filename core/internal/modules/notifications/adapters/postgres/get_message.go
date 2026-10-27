package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) GetMessage(ctx context.Context, id,
	recipient domain.ID,
) (domain.Message, error) {
	if !store.usable(ctx, false) || !id.Valid() || !recipient.Valid() {
		return domain.Message{}, ErrInvalid
	}
	value, err := scanMessage(store.queryRow(ctx, getMessageSQL,
		id.String(), recipient.String()))
	if store.notFound(err) {
		return domain.Message{}, ErrNotFound
	}
	return value, err
}
