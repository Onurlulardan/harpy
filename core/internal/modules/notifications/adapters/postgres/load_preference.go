package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) LoadPreference(ctx context.Context, recipient domain.ID,
	category domain.Category, channel domain.Channel,
) (domain.Preference, bool, error) {
	if !store.usable(ctx, false) || !recipient.Valid() || !category.Valid() ||
		!channel.Valid() {
		return domain.Preference{}, false, ErrInvalid
	}
	value, err := scanPreference(store.queryRow(ctx, getPreferenceSQL,
		recipient.String(), string(category), string(channel)))
	if store.notFound(err) {
		return domain.Preference{}, false, nil
	}
	return value, err == nil, err
}
