package lifecycle

import (
	"context"
	"time"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func demoOperation(kind catalogdomain.OperationKind) bool {
	return kind == catalogdomain.OperationDemoInstall ||
		kind == catalogdomain.OperationDemoRemove
}

func (application *Application) stageDemoApplying(
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
	if err != nil || plan.OperationID != operationID ||
		!demoPlanMatchesKind(plan, operation.Kind) {
		return ResumeResult{}, ErrUnavailable
	}
	components, err := demoStore.DemoComponents(ctx, plan.InstallationID)
	if err != nil {
		return ResumeResult{}, err
	}
	status, err := compileDemoStatus(plan.InstallationID, components)
	if err != nil {
		return ResumeResult{}, err
	}
	if operation.Kind == catalogdomain.OperationDemoRemove {
		if status.State != DemoStateInstalled {
			return ResumeResult{}, ErrDemoPlanStale
		}
		current, planErr := application.planCurrentDemoRemoval(ctx, scope, components)
		if planErr != nil || current.PlanSHA256 != plan.ExpectedPlanSHA256 ||
			!sameDemoRemovalComponents(current.Components, plan.Components) {
			return ResumeResult{}, ErrDemoPlanStale
		}
	} else if status.State != DemoStateAbsent {
		return ResumeResult{}, ErrConflict
	}
	for _, planned := range plan.Components {
		component, found := findDemoComponent(components, planned.ComponentID)
		if !found {
			return ResumeResult{}, ErrUnavailable
		}
		if err = application.applyDemoComponent(
			ctx, scope, demoStore, operation, plan, component, planned,
		); err != nil {
			return ResumeResult{}, err
		}
	}
	return application.advanceOperation(
		ctx, store, operation, catalogdomain.OperationVerifying,
	)
}

func (application *Application) applyDemoComponent(
	ctx context.Context,
	scope platformtx.TenantScope,
	store DemoStore,
	operation catalogdomain.Operation,
	plan DemoOperationPlan,
	component DemoComponent,
	planned DemoRemovalComponent,
) error {
	operationID := uuid.MustParse(operation.ID.String())
	key := operation.ID.String() + ":" + component.ComponentID.String()
	command := demoOwnerCommand(component, operationID, key,
		plan.ExpectedPlanSHA256, planned.RecordIDs)
	action := DemoActionInstall
	result, err := application.demo.Install(ctx, scope, command)
	if operation.Kind == catalogdomain.OperationDemoRemove {
		action = DemoActionRemove
		result, err = application.demo.Remove(ctx, scope, command)
	}
	if err != nil || result.SourceRunID != component.SourceRunID ||
		result.CompletedAt.Before(component.UpdatedAt) ||
		!validDemoOwnerMutation(result, action, planned.RecordIDs) {
		return ErrUnavailable
	}
	recordIDs, _ := canonicalDemoRecordIDs(result.RecordIDs)
	updated, err := store.ApplyDemoResult(ctx, DemoResultMutation{
		OperationID: operationID, ReceiptID: result.ReceiptID,
		InstallationID: component.InstallationID, ComponentID: component.ComponentID,
		Action: action, ContentSHA256: component.ContentSHA256,
		IdempotencyKey: key, SourceRunID: result.SourceRunID,
		RecordIDs: recordIDs, ExpectedRevision: component.Revision,
		CompletedAt: result.CompletedAt,
	})
	wantState := DemoStateInstalled
	if action == DemoActionRemove {
		wantState = DemoStateAbsent
	}
	if err != nil || updated.State != wantState ||
		updated.Revision != component.Revision+1 {
		return ErrUnavailable
	}
	return nil
}

func validDemoOwnerMutation(
	result DemoOwnerResult,
	action string,
	expected []uuid.UUID,
) bool {
	ids, err := canonicalDemoRecordIDs(result.RecordIDs)
	if err != nil || result.ReceiptID.Version() != 7 || result.SourceRunID.Version() != 7 ||
		result.CompletedAt.IsZero() || result.CompletedAt.Location() != time.UTC ||
		result.CompletedAt.Nanosecond()%1000 != 0 {
		return false
	}
	if action == DemoActionInstall {
		return result.Installed && len(ids) > 0
	}
	return !result.Installed && sameUUIDs(ids, expected)
}
