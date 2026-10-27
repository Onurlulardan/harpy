package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type startInput struct {
	OperationID        uuid.UUID
	Kind               catalogdomain.OperationKind
	PackageID          uuid.UUID
	Version            string
	IdempotencyKey     string
	Installation       catalogdomain.Installation
	PreviousOperation  *uuid.UUID
	PreviousState      *catalogdomain.InstallationState
	MutateInstallation bool
	RequestDocument    any
	PurgeAuthority     *PurgeAuthority
}

func (application *Application) start(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	identity ScopeIdentity,
	input startInput,
) (OperationHandle, error) {
	manifest, err := application.loadManifest(ctx, scope, input.PackageID, input.Version)
	if err != nil {
		return OperationHandle{}, err
	}
	var plan catalogdomain.Plan
	if provisioningOperation(input.Kind) {
		plan, err = application.buildProvisioningPlan(
			ctx, scope, store, manifest,
		)
	} else {
		plan, err = application.buildPlan(ctx, scope, store, manifest)
	}
	if err != nil {
		return OperationHandle{}, err
	}
	operationID := input.OperationID
	if operationID == uuid.Nil {
		operationID, err = application.nextIdentity()
		if err != nil {
			return OperationHandle{}, err
		}
	}
	now, err := application.now()
	if err != nil {
		return OperationHandle{}, err
	}
	stableOperationID, err := stableIdentity(operationID)
	if err != nil {
		return OperationHandle{}, err
	}
	operation, err := catalogdomain.NewOperation(stableOperationID, input.Kind, plan, now)
	if err != nil {
		return OperationHandle{}, ErrInvalid
	}
	operation.RetryOf, err = cloneStableID(input.PreviousOperation)
	if err != nil {
		return OperationHandle{}, err
	}
	digest, err := requestDigest(input.RequestDocument)
	if err != nil {
		return OperationHandle{}, err
	}
	handle, err := CreateOperation(ctx, store, StartMutation{
		Installation: input.Installation, Operation: operation,
		PreviousOperation: input.PreviousOperation, PreviousState: input.PreviousState,
		ActorID: identity.ActorID, IdempotencyKey: input.IdempotencyKey,
		RequestSHA256: digest, OccurredAt: now, MutateInstallation: input.MutateInstallation,
		PurgeAuthority: input.PurgeAuthority,
	})
	if err != nil {
		return OperationHandle{}, err
	}
	if handle.Operation.ID != stableOperationID && !handle.Replayed || !handle.Operation.Plan.Valid() {
		return OperationHandle{}, ErrUnavailable
	}
	if !handle.Replayed {
		if err := application.jobs.Dispatch(ctx, scope, handle.Operation); err != nil {
			return OperationHandle{}, err
		}
	}
	return handle, nil
}

func provisioningOperation(kind catalogdomain.OperationKind) bool {
	return kind == catalogdomain.OperationInstall ||
		kind == catalogdomain.OperationUpgrade ||
		kind == catalogdomain.OperationRetry
}

func cloneStableID(value *uuid.UUID) (*packagecontract.StableID, error) {
	if value == nil {
		return nil, nil
	}
	identity, err := stableIdentity(*value)
	return &identity, err
}
