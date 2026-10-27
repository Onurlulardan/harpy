package publication

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) List(
	ctx context.Context, scope platformtx.TenantScope, limit int, after *uuid.UUID,
) (Page, error) {
	if limit < 1 || limit > 100 || (after != nil && !validUUIDv7(*after)) {
		return Page{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, OperationList, PermissionRead)
	if err != nil {
		return Page{}, err
	}
	result, err := store.List(ctx, limit, after)
	if err != nil || len(result.Items) > limit {
		return Page{}, ErrUnavailable
	}
	for _, item := range result.Items {
		if validateSnapshot(item) != nil {
			return Page{}, ErrUnavailable
		}
	}
	if result.After != nil && !validUUIDv7(*result.After) {
		return Page{}, ErrUnavailable
	}
	return result, nil
}
