package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const demoComponentContract = "catalog.demo-data.v1"

type demoStart struct {
	Action, ExpectedPlanSHA256 string
	Installation               catalogdomain.Installation
	Components                 []DemoRemovalComponent
	ResolveComponents          func() ([]DemoRemovalComponent, error)
}

func (application *Application) startDemoOperation(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	identity ScopeIdentity,
	input demoStart,
	idempotencyKey string,
) (OperationHandle, error) {
	packageID, err := uuidFromCanonical(input.Installation.PackageID.String())
	if err != nil {
		return OperationHandle{}, err
	}
	manifest, err := application.loadManifest(
		ctx, scope, packageID, input.Installation.ActiveVersion.String(),
	)
	if err != nil {
		return OperationHandle{}, err
	}
	fullPlan, err := application.buildPlan(ctx, scope, store, manifest)
	if err != nil {
		return OperationHandle{}, err
	}
	demoPins := make([]catalogdomain.ComponentPin, 0, len(fullPlan.Components))
	for _, pin := range fullPlan.Components {
		if pin.Owner == "catalog" && pin.Contract == demoComponentContract {
			demoPins = append(demoPins, pin)
		}
	}
	plan, err := catalogdomain.NewSealedPlan(
		fullPlan.PackageID, fullPlan.Version, fullPlan.Slug,
		fullPlan.PackageSHA256, fullPlan.Dependencies, demoPins, nil,
	)
	if err != nil {
		return OperationHandle{}, ErrConflict
	}
	operationID, err := application.nextIdentity()
	if err != nil {
		return OperationHandle{}, err
	}
	now, err := application.now()
	if err != nil {
		return OperationHandle{}, err
	}
	stableID, err := stableIdentity(operationID)
	if err != nil {
		return OperationHandle{}, err
	}
	kind := catalogdomain.OperationDemoInstall
	if input.Action == DemoActionRemove {
		kind = catalogdomain.OperationDemoRemove
	}
	operation, err := catalogdomain.NewOperation(stableID, kind, plan, now)
	if err != nil {
		return OperationHandle{}, ErrInvalid
	}
	digest, err := requestDigest(struct {
		Action, InstallationID, ExpectedPlanSHA256 string
	}{input.Action, input.Installation.ID.String(), input.ExpectedPlanSHA256})
	if err != nil {
		return OperationHandle{}, err
	}
	handle, err := CreateOperation(ctx, store, StartMutation{
		Installation: input.Installation, Operation: operation,
		ActorID: identity.ActorID, IdempotencyKey: idempotencyKey,
		RequestSHA256: digest, OccurredAt: now,
	})
	if err != nil {
		return OperationHandle{}, err
	}
	demoStore, ok := store.(DemoStore)
	if !ok || demoStore == nil {
		return OperationHandle{}, ErrUnavailable
	}
	if handle.Replayed {
		stored, readErr := demoStore.ReadDemoOperationPlan(
			ctx, uuid.MustParse(handle.Operation.ID.String()),
		)
		if readErr != nil || stored.Action != input.Action ||
			stored.ExpectedPlanSHA256 != input.ExpectedPlanSHA256 {
			return OperationHandle{}, ErrConflict
		}
		return handle, nil
	}
	components := input.Components
	if input.ResolveComponents != nil {
		components, err = input.ResolveComponents()
		if err != nil {
			return OperationHandle{}, err
		}
	}
	stored, err := demoStore.RecordDemoOperationPlan(ctx, DemoOperationPlan{
		OperationID: operationID, InstallationID: uuid.MustParse(input.Installation.ID.String()),
		Action: input.Action, ExpectedPlanSHA256: input.ExpectedPlanSHA256,
		Components: components, CreatedAt: now,
	})
	if err != nil || stored.OperationID != operationID || stored.Action != input.Action {
		return OperationHandle{}, ErrUnavailable
	}
	if err = application.jobs.Dispatch(ctx, scope, handle.Operation); err != nil {
		return OperationHandle{}, err
	}
	return handle, nil
}
