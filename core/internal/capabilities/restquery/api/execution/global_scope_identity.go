package execution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type globalScopeIdentity struct {
	scope         transaction.GlobalScope
	environmentID uuid.UUID
	actorID       uuid.UUID
	mode          transaction.ScopeMode
	operationID   string
}

func captureGlobalScopeIdentity(
	ctx context.Context,
	scope transaction.GlobalScope,
) (globalScopeIdentity, error) {
	if ctx == nil || ctx.Err() != nil || transaction.NetworkAllowed(ctx) ||
		!transaction.ContextOwnsGlobalScope(ctx, scope) ||
		transaction.RequireActiveGlobal(scope, transaction.SnapshotQuery) != nil {
		return globalScopeIdentity{}, errExecutionContract
	}
	identity := observeGlobalScope(scope)
	if !identity.valid() {
		return globalScopeIdentity{}, errExecutionContract
	}
	return identity, nil
}

func observeGlobalScope(scope transaction.GlobalScope) globalScopeIdentity {
	return globalScopeIdentity{
		scope: scope, environmentID: scope.EnvironmentID(), actorID: scope.ActorID(),
		mode: scope.Mode(), operationID: scope.OperationID(),
	}
}

func (identity globalScopeIdentity) valid() bool {
	return identity.environmentID != uuid.Nil &&
		identity.mode == transaction.SnapshotQuery && identity.operationID != ""
}

func (identity globalScopeIdentity) matches(other globalScopeIdentity) bool {
	return identity.scope == other.scope && identity.environmentID == other.environmentID &&
		identity.actorID == other.actorID && identity.mode == other.mode &&
		identity.operationID == other.operationID
}

func verifyCapturedGlobalScope(ctx context.Context, identity globalScopeIdentity) error {
	if ctx == nil || ctx.Err() != nil || transaction.NetworkAllowed(ctx) ||
		!transaction.ContextOwnsGlobalScope(ctx, identity.scope) ||
		transaction.RequireActiveGlobal(identity.scope, transaction.SnapshotQuery) != nil ||
		!identity.matches(observeGlobalScope(identity.scope)) {
		return errExecutionContract
	}
	return nil
}
