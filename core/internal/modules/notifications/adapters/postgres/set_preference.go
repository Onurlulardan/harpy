package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) SetPreference(ctx context.Context,
	value domain.Preference, expected int64,
) error {
	if !store.usable(ctx, true) || !value.Valid() || expected < 0 ||
		value.Revision != expected+1 {
		return ErrInvalid
	}
	timezone, start, end := quietArguments(value.QuietHours)
	var revision int64
	err := store.queryRow(ctx, putPreferenceSQL, value.RecipientID.String(),
		string(value.Category), string(value.Channel), value.Enabled, value.Explicit,
		value.VerifiedTarget, timezone, start, end, value.Revision, expected).Scan(&revision)
	if store.notFound(err) {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	if revision != value.Revision {
		return ErrConflict
	}
	return nil
}
