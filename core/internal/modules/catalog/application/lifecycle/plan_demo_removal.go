package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) PlanDemoRemoval(
	ctx context.Context,
	scope platformtx.TenantScope,
	installationID uuid.UUID,
) (DemoRemovalPlan, error) {
	if installationID.Version() != 7 {
		return DemoRemovalPlan{}, ErrInvalid
	}
	store, demoStore, _, err := application.bindDemoStore(
		ctx, scope, ScopeSnapshot, operationDemoPlan, ActionDemoManage,
	)
	if err != nil {
		return DemoRemovalPlan{}, err
	}
	if err = application.requireDemoOwner(ctx, scope); err != nil {
		return DemoRemovalPlan{}, err
	}
	reader, ok := store.(InstallationByIDStore)
	if !ok || reader == nil {
		return DemoRemovalPlan{}, ErrUnavailable
	}
	installation, err := reader.InstallationByID(ctx, installationID)
	if err != nil {
		return DemoRemovalPlan{}, ErrNotFound
	}
	if installation.State != catalogdomain.InstallationActive &&
		installation.State != catalogdomain.InstallationDisabled {
		return DemoRemovalPlan{}, ErrConflict
	}
	components, err := demoStore.DemoComponents(ctx, installationID)
	if err != nil {
		return DemoRemovalPlan{}, err
	}
	status, err := compileDemoStatus(installationID, components)
	if err != nil || status.State != DemoStateInstalled {
		return DemoRemovalPlan{}, ErrConflict
	}
	return application.planCurrentDemoRemoval(ctx, scope, components)
}

func (application *Application) planCurrentDemoRemoval(
	ctx context.Context,
	scope platformtx.TenantScope,
	components []DemoComponent,
) (DemoRemovalPlan, error) {
	plans := make([]DemoRemovalComponent, 0, len(components))
	for _, component := range components {
		result, err := application.demo.PlanRemoval(
			ctx, scope, demoOwnerCommand(component, uuid.Nil, "", "", nil),
		)
		if err != nil || !result.Installed || result.SourceRunID != component.SourceRunID {
			return DemoRemovalPlan{}, ErrUnavailable
		}
		ids, err := canonicalDemoRecordIDs(result.RecordIDs)
		if err != nil || len(ids) == 0 {
			return DemoRemovalPlan{}, ErrUnavailable
		}
		plans = append(plans, DemoRemovalComponent{
			ComponentID: component.ComponentID,
			SourceRunID: component.SourceRunID,
			RecordIDs:   ids,
		})
	}
	return sealDemoRemovalPlan(components[0].InstallationID, plans)
}

func demoOwnerCommand(
	component DemoComponent,
	operationID uuid.UUID,
	idempotencyKey, expectedPlan string,
	recordIDs []uuid.UUID,
) DemoOwnerCommand {
	return DemoOwnerCommand{
		OperationID: operationID, InstallationID: component.InstallationID,
		PackageID: component.PackageID, ComponentID: component.ComponentID,
		ReleaseID: component.ReleaseID, SourceRunID: component.SourceRunID,
		PackageVersion: component.PackageVersion, ContentSHA256: component.ContentSHA256,
		IdempotencyKey: idempotencyKey, ExpectedPlanSHA256: expectedPlan,
		Document:  append([]byte(nil), component.Document...),
		RecordIDs: append([]uuid.UUID(nil), recordIDs...),
	}
}
