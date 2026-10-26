package catalogcomponentowner

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (owner *Owner) RestoreComponent(
	ctx context.Context,
	scope platformtx.TenantScope,
	command componentowner.LifecycleCommand,
) (componentowner.OwnerResult, error) {
	store, packageID, err := owner.lifecycleStore(
		ctx,
		scope,
		command.TenantID,
		command.PackageID,
		command.Component,
		command.Valid() && command.Action == "restore",
	)
	if err != nil {
		return componentowner.OwnerResult{}, err
	}
	receipt, err := owner.BaseOwner.Restore(
		ctx,
		scope,
		legacyLifecycleCommand(command),
	)
	if err != nil {
		return componentowner.OwnerResult{}, err
	}
	binding, err := store.Binding(ctx, command.Component, packageID)
	if err != nil {
		return componentowner.OwnerResult{}, err
	}
	result := lifecycleResult(command, receipt, binding)
	if !result.ValidForLifecycle(command) {
		return componentowner.OwnerResult{}, unavailableOwnerResult()
	}
	return result, nil
}
