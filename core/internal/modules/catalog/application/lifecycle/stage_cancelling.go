package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) stageCancelling(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	operation catalogdomain.Operation,
) (ResumeResult, error) {
	if operation.DestructiveStarted != nil {
		blocked, err := operation.Advance(catalogdomain.OperationBlocked)
		if err != nil {
			return ResumeResult{}, ErrOutcomeUnknown
		}
		blocked.FailureCode = "CATALOG_OUTCOME_UNKNOWN"
		stored, err := ChangeOperation(ctx, store, OperationMutation{
			Operation: blocked, ExpectedRevision: operation.Revision,
			FailureCode: blocked.FailureCode,
		})
		if err != nil {
			return ResumeResult{}, err
		}
		return ResumeResult{Operation: stored}, ErrOutcomeUnknown
	}
	if _, err := application.applyOwnerSteps(ctx, scope, store, operation, "cleanup", 0, true, 0); err != nil {
		return ResumeResult{}, err
	}
	return application.advanceOperation(ctx, store, operation, catalogdomain.OperationCancelled)
}
