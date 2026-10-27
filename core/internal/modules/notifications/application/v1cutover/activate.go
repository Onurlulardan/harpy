package v1cutover

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

func (application *Application) ActivateV2(ctx context.Context, scope api.Scope,
	expected Authority,
) (Authority, error) {
	store, err := application.store(ctx, scope)
	if err != nil || !expected.Valid() || expected.Writer != "v1" ||
		expected.State != "fenced" {
		return Authority{}, ErrInvalid
	}
	return store.ActivateV2(ctx, expected)
}
