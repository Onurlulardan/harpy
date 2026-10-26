package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const snapshotOperationID = "restquery.execute"

func validateScope(
	ctx context.Context,
	scope transaction.TenantScope,
	batch compiler.RecordBatch,
) error {
	if ctx == nil || ctx.Err() != nil || transaction.NetworkAllowed(ctx) ||
		!transaction.ContextOwnsScope(ctx, scope) ||
		transaction.RequireActive(scope, transaction.SnapshotQuery) != nil ||
		scope.Mode() != transaction.SnapshotQuery ||
		scope.OperationID() != snapshotOperationID {
		return ErrContract
	}
	tenant, tenantErr := batch.TenantID()
	actor, actorErr := batch.ActorID()
	if tenantErr != nil || actorErr != nil || tenant != scope.TenantID() ||
		actor != scope.ActorID() {
		return ErrContract
	}
	return nil
}
