package transaction

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func (coordinator *Coordinator) InProviderTenant(
	ctx context.Context,
	request ProviderScopeRequest,
	execute func(context.Context, ProviderTenantScope) error,
) error {
	if coordinator == nil || coordinator.pool == nil || ctx == nil || execute == nil ||
		!request.valid() {
		return ErrInvalidProviderScopeRequest
	}
	if nestedScope(ctx) {
		return ErrNestedScope
	}
	database, err := coordinator.pool.BeginRuntimeTransaction(ctx, postgres.TransactionCommand)
	if err != nil {
		return ErrScopeDatabase
	}
	target, err, resolverPanic := resolveProviderTarget(ctx, database, request)
	if resolverPanic != nil {
		_ = database.Rollback()
		panic(resolverPanic)
	}
	if err != nil {
		return rollbackAfter(database, err)
	}
	scope, err := prepareProviderScope(ctx, database, request.Lookup, target)
	if err != nil {
		return rollbackAfter(database, err)
	}
	callbackErr, panicValue := callProviderScope(ctx, scope, execute)
	scope.state.active.Store(false)
	sealErr := database.SealBindings()
	if panicValue != nil {
		_ = database.Rollback()
		panic(panicValue)
	}
	if sealErr != nil {
		callbackErr = errors.Join(callbackErr, ErrScopeDatabase)
	}
	if callbackErr != nil {
		return rollbackAfter(database, callbackErr)
	}
	if err := ctx.Err(); err != nil {
		return rollbackAfter(database, err)
	}
	err = database.Commit(ctx)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, postgres.ErrCommitRolledBack):
		return ErrCommitRejected
	default:
		return AmbiguousCommitError{}
	}
}
