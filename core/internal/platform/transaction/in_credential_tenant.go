package transaction

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func (coordinator *Coordinator) InCredentialTenant(
	ctx context.Context,
	request CredentialScopeRequest,
	execute func(context.Context, CredentialTenantScope) error,
) error {
	if coordinator == nil || coordinator.pool == nil || ctx == nil || execute == nil ||
		!request.valid() {
		return ErrInvalidCredentialScopeRequest
	}
	if nestedScope(ctx) {
		return ErrNestedScope
	}
	database, err := coordinator.pool.BeginRuntimeTransaction(ctx, postgres.TransactionCommand)
	if err != nil {
		return ErrScopeDatabase
	}
	target, err, resolverPanic := resolveCredentialTarget(ctx, database, request)
	if resolverPanic != nil {
		_ = database.Rollback()
		panic(resolverPanic)
	}
	if err != nil {
		return rollbackAfter(database, err)
	}
	scope, err := prepareCredentialScope(ctx, database, target)
	if err != nil {
		return rollbackAfter(database, err)
	}
	callbackErr, panicValue := callCredentialScope(ctx, scope, execute)
	scope.state.active.Store(false)
	sealErr := database.SealBindings()
	if panicValue != nil {
		_ = database.Rollback()
		panic(panicValue)
	}
	if sealErr != nil {
		cause := ErrScopeDatabase
		if callbackErr != nil {
			cause = errors.Join(callbackErr, cause)
		}
		return rollbackAfter(database, cause)
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
