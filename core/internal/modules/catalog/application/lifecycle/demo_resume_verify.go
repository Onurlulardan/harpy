package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) stageDemoVerifying(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	operation catalogdomain.Operation,
) (ResumeResult, error) {
	demoStore, ok := store.(DemoStore)
	if !ok || demoStore == nil || application.requireDemoOwner(ctx, scope) != nil {
		return ResumeResult{}, ErrUnavailable
	}
	operationID := uuid.MustParse(operation.ID.String())
	plan, err := demoStore.ReadDemoOperationPlan(ctx, operationID)
	if err != nil || !demoPlanMatchesKind(plan, operation.Kind) {
		return ResumeResult{}, ErrUnavailable
	}
	components, err := demoStore.DemoComponents(ctx, plan.InstallationID)
	if err != nil {
		return ResumeResult{}, err
	}
	status, err := compileDemoStatus(plan.InstallationID, components)
	wantInstalled := operation.Kind == catalogdomain.OperationDemoInstall
	wantState := DemoStateAbsent
	if wantInstalled {
		wantState = DemoStateInstalled
	}
	if err != nil || status.State != wantState || len(components) != len(plan.Components) {
		return ResumeResult{}, ErrOwnerReceiptStale
	}
	for _, planned := range plan.Components {
		component, found := findDemoComponent(components, planned.ComponentID)
		if !found {
			return ResumeResult{}, ErrOwnerReceiptStale
		}
		command := demoOwnerCommand(component, operationID,
			operation.ID.String()+":"+component.ComponentID.String()+":inspect",
			plan.ExpectedPlanSHA256, component.RecordIDs)
		result, inspectErr := application.demo.Inspect(ctx, scope, command)
		if inspectErr != nil || result.Installed != wantInstalled ||
			result.SourceRunID != component.SourceRunID ||
			!sameUUIDs(result.RecordIDs, component.RecordIDs) {
			return ResumeResult{}, ErrOwnerReceiptStale
		}
	}
	return application.advanceOperation(
		ctx, store, operation, catalogdomain.OperationActivating,
	)
}

func (application *Application) stageDemoActivating(
	ctx context.Context,
	store Store,
	operation catalogdomain.Operation,
) (ResumeResult, error) {
	return application.advanceOperation(
		ctx, store, operation, catalogdomain.OperationSucceeded,
	)
}

func demoPlanMatchesKind(
	plan DemoOperationPlan,
	kind catalogdomain.OperationKind,
) bool {
	return plan.OperationID.Version() == 7 && plan.InstallationID.Version() == 7 &&
		len(plan.Components) > 0 &&
		(kind == catalogdomain.OperationDemoInstall && plan.Action == DemoActionInstall ||
			kind == catalogdomain.OperationDemoRemove && plan.Action == DemoActionRemove)
}

func findDemoComponent(
	components []DemoComponent,
	componentID uuid.UUID,
) (DemoComponent, bool) {
	for _, component := range components {
		if component.ComponentID == componentID {
			return component, true
		}
	}
	return DemoComponent{}, false
}

func sameDemoRemovalComponents(
	left, right []DemoRemovalComponent,
) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].ComponentID != right[index].ComponentID ||
			left[index].SourceRunID != right[index].SourceRunID ||
			!sameUUIDs(left[index].RecordIDs, right[index].RecordIDs) {
			return false
		}
	}
	return true
}

func sameUUIDs(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
