package dashboard

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) ReadDefault(
	ctx context.Context, scope platformtx.TenantScope, surface string,
) (Snapshot, error) {
	if !surfacePattern.MatchString(surface) {
		return Snapshot{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, OperationRead, PermissionRead)
	if err != nil {
		return Snapshot{}, err
	}
	result, err := store.ReadDefault(ctx, surface)
	if err != nil {
		return Snapshot{}, err
	}
	if validateSnapshot(result) != nil {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
