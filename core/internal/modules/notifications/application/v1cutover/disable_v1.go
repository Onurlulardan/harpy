package v1cutover

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

func (application *Application) DisableV1(ctx context.Context, scope api.Scope,
	expected Authority, evidence Evidence,
) (Authority, error) {
	store, err := application.store(ctx, scope)
	if err != nil || !expected.Valid() || expected.Writer != "v1" ||
		expected.State != "quiescing" || evidence.ID == [16]byte{} ||
		!evidence.Inventory.Zero() {
		return Authority{}, ErrInvalid
	}
	return store.DisableV1(ctx, expected, evidence)
}
