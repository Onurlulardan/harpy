package lifecycle

import (
	"context"
	"time"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ResumeResult struct {
	Operation catalogdomain.Operation
	NextAt    time.Time
}

func IsTerminalResume(operation Operation) bool {
	switch operation.State {
	case catalogdomain.OperationSucceeded, catalogdomain.OperationBlocked,
		catalogdomain.OperationFailed, catalogdomain.OperationCancelled:
		return true
	default:
		return false
	}
}

func (application *Application) Resume(
	ctx context.Context,
	scope platformtx.TenantScope,
	operationID uuid.UUID,
) (ResumeResult, error) {
	if operationID == uuid.Nil {
		return ResumeResult{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeCommand, operationResume, ActionResume)
	if err != nil {
		return ResumeResult{}, err
	}
	operation, err := store.ReadOperation(ctx, operationID)
	if err != nil {
		return ResumeResult{}, ErrNotFound
	}
	switch operation.State {
	case catalogdomain.OperationPlanned:
		return application.advanceOperation(ctx, store, operation, catalogdomain.OperationApplying)
	case catalogdomain.OperationWaiting:
		return application.stageWaiting(ctx, scope, store, operation)
	case catalogdomain.OperationApplying:
		return application.stageApplying(ctx, scope, store, operation)
	case catalogdomain.OperationVerifying:
		return application.stageVerifying(ctx, scope, store, operation)
	case catalogdomain.OperationActivating:
		return application.stageActivating(ctx, scope, store, operation)
	case catalogdomain.OperationCancelling:
		return application.stageCancelling(ctx, scope, store, operation)
	case catalogdomain.OperationSucceeded, catalogdomain.OperationCancelled:
		return ResumeResult{Operation: operation}, nil
	default:
		return ResumeResult{}, ErrConflict
	}
}

func (application *Application) advanceOperation(
	ctx context.Context,
	store Store,
	current catalogdomain.Operation,
	nextState catalogdomain.OperationState,
) (ResumeResult, error) {
	next, err := current.Advance(nextState)
	if err != nil {
		return ResumeResult{}, ErrConflict
	}
	stored, err := ChangeOperation(ctx, store, OperationMutation{
		Operation: next, ExpectedRevision: current.Revision,
	})
	if err != nil || stored.Revision != next.Revision {
		return ResumeResult{}, ErrUnavailable
	}
	return ResumeResult{Operation: stored}, nil
}
