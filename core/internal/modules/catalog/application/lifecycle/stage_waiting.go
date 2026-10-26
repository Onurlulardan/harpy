package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) stageWaiting(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	operation catalogdomain.Operation,
) (ResumeResult, error) {
	if operation.Kind == catalogdomain.OperationPurge {
		authority, now, err := application.revalidatePurgeAuthority(
			ctx, scope, store, operation,
		)
		if err != nil {
			return ResumeResult{}, err
		}
		if now.Before(authority.CoolingUntil) {
			return ResumeResult{Operation: operation, NextAt: authority.CoolingUntil}, nil
		}
	}
	return application.advanceOperation(ctx, store, operation, catalogdomain.OperationApplying)
}
