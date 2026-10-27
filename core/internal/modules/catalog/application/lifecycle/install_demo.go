package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) InstallDemo(
	ctx context.Context,
	scope platformtx.TenantScope,
	request DemoInstallRequest,
) (OperationHandle, error) {
	if request.InstallationID.Version() != 7 ||
		!validIdempotency(request.IdempotencyKey) {
		return OperationHandle{}, ErrInvalid
	}
	store, demoStore, identity, err := application.bindDemoStore(
		ctx, scope, ScopeCommand, operationDemoInstall, ActionDemoManage,
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
		Action: DemoActionInstall, Installation: installation,
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
			if statusErr != nil || status.State != DemoStateAbsent {
				return nil, ErrConflict
			}
			return demoComponentReferences(components), nil
		},
	}, request.IdempotencyKey)
}

func demoComponentReferences(components []DemoComponent) []DemoRemovalComponent {
	result := make([]DemoRemovalComponent, len(components))
	for index, component := range components {
		result[index] = DemoRemovalComponent{
			ComponentID: component.ComponentID,
			SourceRunID: component.SourceRunID,
		}
	}
	return result
}
