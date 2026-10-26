package v1cutover

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"
)

func (store *Store) CreateAuthority(ctx context.Context, origin app.Origin) (app.Authority, error) {
	if !store.usable(ctx) || origin != app.OriginFresh && origin != app.OriginV1Adopted {
		return app.Authority{}, app.ErrInvalid
	}
	value, err := scanAuthority(store.queryRow(ctx, bootstrapSQL, string(origin)))
	if err == nil {
		return value, nil
	}
	value, err = scanAuthority(store.queryRow(ctx, getAuthoritySQL))
	if err != nil || origin == app.OriginFresh && value.Writer != "v2" ||
		origin == app.OriginV1Adopted && value.Writer != "v1" && value.Writer != "v2" {
		return app.Authority{}, app.ErrConflict
	}
	return value, nil
}
