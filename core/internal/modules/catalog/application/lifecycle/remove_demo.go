package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) RemoveDemo(
	ctx context.Context,
	scope platformtx.TenantScope,
	request DemoRemoveRequest,
) (OperationHandle, error) {
	if request.InstallationID.Version() != 7 ||
		!sha256Pattern.MatchString(request.ExpectedPlanSHA256) ||
		!validIdempotency(request.IdempotencyKey) {
		return OperationHandle{}, ErrInvalid
	}
	store, demoStore, identity, err := application.bindDemoStore(
		ctx, scope, ScopeCommand, operationDemoRemove, ActionDemoManage,
	)
	if err != nil {
		return OperationHandle{}, err
	}
	if err = application.requireDemoOwner(ctx, scope); err != nil {
		return OperationHandle{}, err
	}
	reader, ok := store.(InstallationByIDStore)
	if !ok || reader == nil {
		return OperationHandle{}, ErrUnavailable
	}
	installation, err := reader.InstallationByID(ctx, request.InstallationID)
	if err != nil {
		return OperationHandle{}, ErrNotFound
	}
	return application.startDemoOperation(ctx, scope, store, identity, demoStart{
		Action: DemoActionRemove, Installation: installation,
		ExpectedPlanSHA256: request.ExpectedPlanSHA256,
		ResolveComponents: func() ([]DemoRemovalComponent, error) {
			if installation.State != catalogdomain.InstallationActive &&
				installation.State != catalogdomain.InstallationDisabled {
				return nil, ErrConflict
			}
			components, loadErr := demoStore.DemoComponents(ctx, request.InstallationID)
			if loadErr != nil {
				return nil, loadErr
			}
			status, statusErr := compileDemoStatus(request.InstallationID, components)
			if statusErr != nil || status.State != DemoStateInstalled {
				return nil, ErrConflict
			}
			plan, planErr := application.planCurrentDemoRemoval(ctx, scope, components)
			if planErr != nil {
				return nil, planErr
			}
			if plan.PlanSHA256 != request.ExpectedPlanSHA256 {
				return nil, ErrDemoPlanStale
			}
			return plan.Components, nil
		},
	}, request.IdempotencyKey)
}
