package v1cutover

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

func (application *Application) PauseV1(ctx context.Context, scope api.Scope,
	expected Authority,
) (Authority, error) {
	store, err := application.store(ctx, scope)
	if err != nil || !expected.Valid() || expected.Writer != "v1" ||
		expected.State != "writable" {
		return Authority{}, ErrInvalid
	}
	return store.PauseV1(ctx, expected)
}
