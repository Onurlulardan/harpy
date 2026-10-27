package v1cutover

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

func (application *Application) RecordEvidence(ctx context.Context, scope api.Scope,
	inventory Inventory,
) (Evidence, error) {
	store, err := application.store(ctx, scope)
	if err != nil || inventory.Validate() != nil {
		return Evidence{}, ErrInvalid
	}
	return store.RecordEvidence(ctx, inventory)
}
