package systemresourcelayouts

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) runCommand(
	ctx context.Context,
	admission publicresource.Admission,
	operation string,
	execute func(context.Context, platformtx.TenantScope) error,
) error {
	if !application.usable(ctx) || execute == nil {
		return ErrContract
	}
	request, err := commandScopeRequest(admission, operation)
	if err != nil {
		return err
	}
	return application.coordinator.InTenant(ctx, request, execute)
}

func (application *Application) runSnapshot(
	ctx context.Context,
	admission publicresource.Admission,
	execute func(context.Context, platformtx.TenantScope) error,
) error {
	if !application.usable(ctx) || execute == nil {
		return ErrContract
	}
	request, err := snapshotScopeRequest(admission)
	if err != nil {
		return err
	}
	return application.coordinator.InTenant(ctx, request, execute)
}
