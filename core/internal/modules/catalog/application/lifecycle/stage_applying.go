package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) stageApplying(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	operation catalogdomain.Operation,
) (ResumeResult, error) {
	if demoOperation(operation.Kind) {
		return application.stageDemoApplying(ctx, scope, store, operation)
	}
	if operation.Kind == catalogdomain.OperationPurge {
		authority, now, err := application.revalidatePurgeAuthority(
			ctx, scope, store, operation,
		)
		if err != nil {
			return ResumeResult{}, err
		}
		if now.Before(authority.CoolingUntil) {
			result, err := application.advanceOperation(ctx, store, operation, catalogdomain.OperationWaiting)
			result.NextAt = authority.CoolingUntil
			return result, err
		}
		operation, retryAt, ready, err := application.preparePurgeDestruction(
			ctx, scope, store, operation, authority, now,
		)
		if err != nil {
			return ResumeResult{}, err
		}
		if !ready {
			result, advanceErr := application.advanceOperation(
				ctx, store, operation, catalogdomain.OperationWaiting,
			)
			result.NextAt = retryAt
			return result, advanceErr
		}
		return application.advanceOperation(
			ctx, store, operation, catalogdomain.OperationVerifying,
		)
	}
	if provisioningOperation(operation.Kind) {
		retryAt, ready, err := application.provisionOwnerSteps(
			ctx, scope, operation,
		)
		if err != nil {
			return ResumeResult{}, err
		}
		if !ready {
			result, err := application.advanceOperation(
				ctx, store, operation, catalogdomain.OperationWaiting,
			)
			result.NextAt = retryAt
			return result, err
		}
	}
	more, err := application.applyOwnerSteps(
		ctx, scope, store, operation, "stage", 0, false, 1,
	)
	if err != nil {
		return ResumeResult{}, err
	}
	if more {
		return application.advanceOperation(
			ctx, store, operation, catalogdomain.OperationWaiting,
		)
	}
	return application.advanceOperation(ctx, store, operation, catalogdomain.OperationVerifying)
}
