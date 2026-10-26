package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (application *Application) bindOperation(ctx context.Context, scope api.Scope,
	permission string,
) (ports.Stores, error) {
	if application == nil || ctx == nil || !validScope(scope) {
		return nil, ErrInvalid
	}
	if err := application.authorizer.AuthorizeOperation(ctx, scope, permission); err != nil {
		return nil, ErrForbidden
	}
	stores, err := application.binder.Bind(ctx, scope)
	if err != nil || stores == nil {
		return nil, ErrUnavailable
	}
	return stores, nil
}
