package lifecycle

import (
	"context"
	_ "embed"
	"time"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/delete_dependency_resolutions.sql
var deleteDependencyResolutionsSQL string

//go:embed sql/tenant/upsert_dependency_resolution.sql
var upsertDependencyResolutionSQL string

//go:embed sql/tenant/upsert_component_binding.sql
var upsertComponentBindingSQL string

func (storage *store) applyOperationPlan(
	ctx context.Context, operationRaw string, installation catalogdomain.Installation,
	occurredAt time.Time,
) error {
	operationID, err := uuid.Parse(operationRaw)
	if err != nil {
		return unavailable()
	}
	operation, err := storage.loadOperation(ctx, operationID)
	if err != nil || operation.Plan.PackageID.String() != installation.PackageID.String() {
		return unavailable()
	}
	installationID, _ := uuid.Parse(installation.ID.String())
	if _, err = storage.execute(ctx, deleteDependencyResolutionsSQL, installationID); err != nil {
		return unavailable()
	}
	if installation.State != catalogdomain.InstallationPurged {
		for _, pin := range operation.Plan.Dependencies {
			if _, err = storage.execute(ctx, upsertDependencyResolutionSQL, installationID,
				pin.PackageID.String(), pin.Version.String(), pin.Required, operation.Plan.SHA256); err != nil {
				return unavailable()
			}
		}
	}
	bindingState, ok := bindingStateFor(installation.State)
	if !ok {
		return unavailable()
	}
	for _, pin := range operation.Plan.Components {
		if _, err = storage.execute(ctx, upsertComponentBindingSQL, installationID,
			pin.ComponentID.String(), pin.Owner, pin.OwnerGeneration,
			pin.Contract, pin.ContentSHA256, bindingState); err != nil {
			return unavailable()
		}
	}
	if installation.State == catalogdomain.InstallationPurged {
		if err = storage.writeTombstone(ctx, operation, installation, occurredAt); err != nil {
			return err
		}
	}
	if operation.Kind == catalogdomain.OperationInstall || operation.Kind == catalogdomain.OperationUpgrade ||
		operation.Kind == catalogdomain.OperationRetry {
		if err = storage.publishOperationMenu(ctx, operation); err != nil {
			return err
		}
	}
	return nil
}

func bindingStateFor(state catalogdomain.InstallationState) (string, bool) {
	switch state {
	case catalogdomain.InstallationActive:
		return "ACTIVE", true
	case catalogdomain.InstallationDisabled:
		return "DISABLED", true
	case catalogdomain.InstallationRetired:
		return "RETIRED", true
	case catalogdomain.InstallationPurged:
		return "PURGED", true
	default:
		return "", false
	}
}
