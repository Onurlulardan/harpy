package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

func (application *Application) CompleteInboxRead(ctx context.Context,
	scope api.Scope,
) (int64, error) {
	stores, err := application.bindOperation(ctx, scope, PermissionInboxUpdate)
	if err != nil {
		return 0, err
	}
	_, recipient, _ := scopeIDs(scope)
	boundary, err := stores.Messages().CurrentInboxRevision(ctx, recipient)
	if err != nil {
		return 0, ErrUnavailable
	}
	if boundary == 0 {
		return 0, nil
	}
	count, err := stores.Messages().CompleteInboxRead(
		ctx, recipient, boundary, scopeActor(scope))
	if err != nil || count < 0 {
		return 0, ErrUnavailable
	}
	return count, nil
}
