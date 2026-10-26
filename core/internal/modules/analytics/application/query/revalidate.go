package query

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Revalidate(
	ctx context.Context, scope platformtx.TenantScope, request PublishRequest,
) (Snapshot, error) {
	if !validPublishRequest(request) {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		OperationRevalidate, PermissionPublish)
	if err != nil {
		return Snapshot{}, err
	}
	return application.publishBound(ctx, scope, store, identity, OperationRevalidate, request)
}
