package query

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ExecutePinnedRequest struct {
	VersionID  uuid.UUID
	Parameters []byte
}

func (application *Application) ExecutePinned(
	ctx context.Context, scope platformtx.TenantScope, request ExecutePinnedRequest,
) (ExecutionResult, error) {
	parameters, err := NewDocument(request.Parameters)
	if !validUUIDv7(request.VersionID) || err != nil {
		return ExecutionResult{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeSnapshot,
		OperationExecute, PermissionExecute)
	if err != nil {
		return ExecutionResult{}, err
	}
	queryID, err := store.ResolveQueryID(ctx, request.VersionID)
	if err != nil {
		return ExecutionResult{}, err
	}
	if !validUUIDv7(queryID) {
		return ExecutionResult{}, ErrUnavailable
	}
	return application.executeBound(ctx, scope, store, identity,
		ExecuteRequest{QueryID: queryID, VersionID: request.VersionID}, parameters)
}
