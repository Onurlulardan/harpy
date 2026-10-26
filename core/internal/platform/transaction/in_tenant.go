package transaction

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func (coordinator *Coordinator) InTenant(
	ctx context.Context,
	request ScopeRequest,
	execute func(context.Context, TenantScope) error,
) error {
	if coordinator == nil || coordinator.pool == nil || ctx == nil || execute == nil ||
		!request.valid() {
		return ErrInvalidScopeRequest
	}
	if nestedScope(ctx) {
		return ErrNestedScope
	}
	transaction, err := coordinator.pool.BeginRuntimeTransaction(ctx, postgresMode(request.Mode))
	if err != nil {
		return ErrScopeDatabase
	}
	scope, err := prepareTenantScope(ctx, transaction, request)
	if err != nil {
		return rollbackAfter(transaction, err)
	}
	callbackErr, panicValue := callScope(ctx, scope, execute)
	scope.state.active.Store(false)
	sealErr := transaction.SealBindings()
	if panicValue != nil {
		_ = transaction.Rollback()
		panic(panicValue)
	}
	if sealErr != nil {
		cause := ErrScopeDatabase
		if callbackErr != nil {
			cause = errors.Join(callbackErr, cause)
		}
		return rollbackAfter(transaction, cause)
	}
	if callbackErr != nil {
		return rollbackAfter(transaction, callbackErr)
	}
	if err := ctx.Err(); err != nil {
		return rollbackAfter(transaction, err)
	}
	err = transaction.Commit(ctx)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, postgres.ErrCommitRolledBack):
		return ErrCommitRejected
	default:
		return AmbiguousCommitError{}
	}
}

func postgresMode(mode ScopeMode) postgres.TransactionMode {
	if mode == SnapshotQuery {
		return postgres.TransactionSnapshotQuery
	}
	return postgres.TransactionCommand
}
