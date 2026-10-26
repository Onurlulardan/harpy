package transaction

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func (coordinator *Coordinator) InGlobal(
	ctx context.Context,
	request GlobalScopeRequest,
	execute func(context.Context, GlobalScope) error,
) error {
	if coordinator == nil || coordinator.pool == nil || ctx == nil || execute == nil ||
		!request.valid() {
		return ErrInvalidGlobalScopeRequest
	}
	if nestedScope(ctx) {
		return ErrNestedScope
	}
	database, err := coordinator.pool.BeginRuntimeTransaction(
		ctx, postgresMode(request.mode()),
	)
	if err != nil {
		return ErrGlobalScopeDatabase
	}
	scope, err := prepareGlobalScope(ctx, database, request)
	if err != nil {
		return rollbackAfter(database, err)
	}
	callbackErr, panicValue := callGlobalScope(ctx, scope, execute)
	scope.state.active.Store(false)
	sealErr := database.SealBindings()
	if panicValue != nil {
		_ = database.Rollback()
		panic(panicValue)
	}
	if sealErr != nil {
		cause := ErrGlobalScopeDatabase
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
