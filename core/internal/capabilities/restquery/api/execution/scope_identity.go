package execution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const snapshotOperationID = "restquery.execute"

type scopeIdentity struct {
	scope       transaction.TenantScope
	tenantID    uuid.UUID
	actorID     uuid.UUID
	mode        transaction.ScopeMode
	operationID string
}

func captureScopeIdentity(
	ctx context.Context,
	scope transaction.TenantScope,
) (scopeIdentity, error) {
	if ctx == nil || ctx.Err() != nil || transaction.NetworkAllowed(ctx) ||
		!transaction.ContextOwnsScope(ctx, scope) ||
		transaction.RequireActive(scope, transaction.SnapshotQuery) != nil {
		return scopeIdentity{}, errExecutionContract
	}
	identity := observeScope(scope)
	if !identity.valid() {
		return scopeIdentity{}, errExecutionContract
	}
	return identity, nil
}

func observeScope(scope transaction.TenantScope) scopeIdentity {
	return scopeIdentity{
		scope: scope, tenantID: scope.TenantID(), actorID: scope.ActorID(),
		mode: scope.Mode(), operationID: scope.OperationID(),
	}
}

func (identity scopeIdentity) valid() bool {
	return identity.tenantID != uuid.Nil && identity.actorID != uuid.Nil &&
		identity.mode == transaction.SnapshotQuery &&
		identity.operationID == snapshotOperationID
}

func (identity scopeIdentity) matches(other scopeIdentity) bool {
	return identity.scope == other.scope && identity.tenantID == other.tenantID &&
		identity.actorID == other.actorID && identity.mode == other.mode &&
		identity.operationID == other.operationID
}

func verifyCapturedScope(ctx context.Context, identity scopeIdentity) error {
	if ctx == nil || ctx.Err() != nil || transaction.NetworkAllowed(ctx) ||
		!transaction.ContextOwnsScope(ctx, identity.scope) ||
		transaction.RequireActive(identity.scope, transaction.SnapshotQuery) != nil ||
		verifyScopeObservation(identity, observeScope(identity.scope)) != nil {
		return errExecutionContract
	}
	return nil
}

func verifyScopeObservation(expected, actual scopeIdentity) error {
	if !expected.matches(actual) {
		return errExecutionContract
	}
	return nil
}
