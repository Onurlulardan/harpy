package lifecycle

import (
	"context"
	"errors"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func ChangeOperation(
	ctx context.Context, store Store, mutation OperationMutation,
) (catalogdomain.Operation, error) {
	if ctx == nil || ctx.Err() != nil || store == nil {
		return catalogdomain.Operation{}, ErrUnavailable
	}
	return store.ChangeOperation(ctx, mutation)
}

func (application *Application) PersistResumeOutcome(
	ctx context.Context,
	scope platformtx.TenantScope,
	operationID uuid.UUID,
	cause error,
) (catalogdomain.Operation, error) {
	state, code, durable := durableResumeFailure(cause)
	if !durable || operationID == uuid.Nil {
		return catalogdomain.Operation{}, cause
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationResume, ActionResume)
	if err != nil {
		return catalogdomain.Operation{}, err
	}
	current, err := store.ReadOperation(ctx, operationID)
	if err != nil {
		return catalogdomain.Operation{}, err
	}
	if current.State == catalogdomain.OperationCancelling {
		return catalogdomain.Operation{}, cause
	}
	next, err := current.Advance(state)
	if err != nil {
		return catalogdomain.Operation{}, ErrConflict
	}
	next.FailureCode = code
	stored, err := ChangeOperation(ctx, store, OperationMutation{
		Operation: next, ExpectedRevision: current.Revision, FailureCode: code,
	})
	if err != nil {
		return catalogdomain.Operation{}, err
	}
	if failErr := application.failInstallingOnDurableStop(
		ctx, store, identity, stored,
	); failErr != nil {
		return catalogdomain.Operation{}, failErr
	}
	return stored, nil
}

func ResumeFailureDurable(cause error) bool {
	_, _, durable := durableResumeFailure(cause)
	return durable
}

func durableResumeFailure(cause error) (catalogdomain.OperationState, string, bool) {
	switch {
	case errors.Is(cause, ErrOwnerNotV2):
		return catalogdomain.OperationBlocked, "CATALOG_OWNER_NOT_V2", true
	case errors.Is(cause, ErrOwnerReceiptStale):
		return catalogdomain.OperationBlocked, "CATALOG_OWNER_RECEIPT_STALE", true
	case errors.Is(cause, ErrDependencyBlocked):
		return catalogdomain.OperationBlocked, "CATALOG_DEPENDENCY_BLOCKED", true
	case errors.Is(cause, ErrBackupBlocked):
		return catalogdomain.OperationBlocked, "CATALOG_BACKUP_BLOCKED", true
	case errors.Is(cause, ErrOutcomeUnknown):
		return catalogdomain.OperationFailed, "CATALOG_OUTCOME_UNKNOWN", true
	case errors.Is(cause, ErrResumeInterrupted),
		errors.Is(cause, context.Canceled),
		errors.Is(cause, context.DeadlineExceeded):
		return catalogdomain.OperationFailed, "CATALOG_RESUME_INTERRUPTED", true
	case errors.Is(cause, ErrDemoPlanStale):
		return catalogdomain.OperationBlocked, "CATALOG_DEMO_PLAN_STALE", true
	default:
		return "", "", false
	}
}
