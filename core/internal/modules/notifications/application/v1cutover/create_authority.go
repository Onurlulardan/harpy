package v1cutover

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

func (application *Application) CreateAuthority(ctx context.Context, scope api.Scope,
	origin Origin,
) (Authority, error) {
	store, err := application.store(ctx, scope)
	if err != nil || origin != OriginFresh && origin != OriginV1Adopted {
		return Authority{}, ErrInvalid
	}
	return store.CreateAuthority(ctx, origin)
}
