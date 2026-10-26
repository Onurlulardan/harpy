package v1cutover

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"
)

func (store *Store) ActivateV2(ctx context.Context,
	expected app.Authority,
) (app.Authority, error) {
	if !store.usable(ctx) || !expected.Valid() || expected.Writer != "v1" ||
		expected.State != "fenced" {
		return app.Authority{}, app.ErrInvalid
	}
	value, err := scanAuthority(store.queryRow(ctx, activateV2SQL,
		expected.Generation, expected.RowVersion))
	if err != nil || value.Writer != "v2" || value.State != "writable" ||
		value.Generation != expected.Generation+1 {
		return app.Authority{}, app.ErrConflict
	}
	return value, nil
}

var _ app.Store = (*Store)(nil)
