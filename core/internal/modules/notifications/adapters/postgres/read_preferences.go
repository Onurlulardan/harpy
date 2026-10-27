package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) ReadPreferences(ctx context.Context,
	recipient domain.ID,
) ([]domain.Preference, error) {
	if !store.usable(ctx, false) || !recipient.Valid() {
		return nil, ErrInvalid
	}
	rows, err := store.query(ctx, listPreferencesSQL, recipient.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.Preference, 0, 18)
	for rows.Next() {
		value, scanErr := scanPreference(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
