package transaction

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func (coordinator *Coordinator) inJobTenant(
	ctx context.Context,
	tenantID uuid.UUID,
	mode postgres.TransactionMode,
	execute func(context.Context, JobTenantScope) error,
) error {
	if coordinator == nil || coordinator.pool == nil || ctx == nil ||
		tenantID == uuid.Nil || execute == nil {
		return ErrInvalidScopeRequest
	}
	if nestedScope(ctx) {
		return ErrNestedScope
	}
	database, err := coordinator.pool.BeginRuntimeTransaction(ctx, mode)
	if err != nil {
		return ErrScopeDatabase
	}
	record, err := database.ResolveTenant(ctx, tenantID)
	if err == nil {
		err = database.BindTenantSearchPath(ctx, record)
	}
	if err == nil {
		err = database.VerifyTenantIdentity(ctx, tenantID, record)
	}
	if err != nil {
		return rollbackAfter(database, ErrScopeDatabase)
	}
	state := &jobScopeState{
		tenantID: tenantID, tenantSlug: record.Slug(), transaction: database,
	}
	state.active.Store(true)
	scope := JobTenantScope{state: state}
	callbackErr, panicValue := callJobScope(ctx, scope, execute)
	state.active.Store(false)
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

func callJobScope(
	ctx context.Context,
	scope JobTenantScope,
	execute func(context.Context, JobTenantScope) error,
) (callbackErr error, panicValue any) {
	defer func() { panicValue = recover() }()
	callbackErr = execute(markJobScopeContext(ctx, scope), scope)
	return callbackErr, nil
}
