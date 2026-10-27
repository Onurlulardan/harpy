package v1adoption

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1adoption"
)

func (store *Store) SourceExists(ctx context.Context) (bool, error) {
	if !store.usable(ctx) {
		return false, app.ErrUnavailable
	}
	var exists bool
	if err := store.queryRow(ctx, sourceExistsSQL).Scan(&exists); err != nil {
		return false, app.ErrUnavailable
	}
	return exists, nil
}
