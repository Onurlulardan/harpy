package v1cutover

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"
)

func (store *Store) PauseV1(ctx context.Context,
	expected app.Authority,
) (app.Authority, error) {
	if !store.usable(ctx) || !expected.Valid() {
		return app.Authority{}, app.ErrInvalid
	}
	value, err := scanAuthority(store.queryRow(ctx, quiesceSQL,
		expected.Generation, expected.RowVersion))
	if err != nil {
		return app.Authority{}, app.ErrConflict
	}
	return value, nil
}
