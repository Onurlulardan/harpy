package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"time"

	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type purgeOwnerStep struct {
	pin   catalogdomain.ComponentPin
	owner componentowner.LifecycleOwner
}

func (application *Application) preflightPurgeOwners(
	ctx context.Context, scope platformtx.TenantScope,
	operation catalogdomain.Operation, authority PurgeAuthority,
) (time.Time, bool, error) {
	steps, err := application.purgeOwnerSteps(operation)
	if err != nil {
		return time.Time{}, false, err
	}
	retryAt, waiting, blocked := time.Time{}, false, false
	for _, step := range steps {
		command := purgeOwnerCommand(operation, authority, step.pin, "preflight")
		result, callErr := step.owner.PreflightComponentPurge(ctx, scope, command)
		if callErr != nil {
			err = errors.Join(err, callErr)
		} else if !result.ValidPreflightFor(command) || !ownerResultExact(step.pin, result) {
			err = errors.Join(err, ErrOwnerReceiptStale)
		} else {
			retryAt, waiting, blocked = collectPurgeState(
				retryAt, waiting, blocked, result,
			)
		}
	}
	return purgeCoordinationResult(retryAt, waiting, blocked, err)
}

func (application *Application) beginPurgeOwners(
	ctx context.Context, scope platformtx.TenantScope,
	operation catalogdomain.Operation, authority PurgeAuthority,
) (time.Time, bool, error) {
	steps, err := application.purgeOwnerSteps(operation)
	if err != nil {
		return time.Time{}, false, err
	}
	retryAt, waiting, blocked := time.Time{}, false, false
	for _, step := range steps {
		command := purgeOwnerCommand(operation, authority, step.pin, "prepare")
		result, callErr := step.owner.BeginComponentPurge(ctx, scope, command)
		if callErr != nil {
			err = errors.Join(err, callErr)
		} else if !result.ValidBeginFor(command) || !ownerResultExact(step.pin, result) {
			err = errors.Join(err, ErrOwnerReceiptStale)
		} else {
			retryAt, waiting, blocked = collectPurgeState(
				retryAt, waiting, blocked, result,
			)
		}
	}
	return purgeCoordinationResult(retryAt, waiting, blocked, err)
}

func (application *Application) finalizePurgeOwners(
	ctx context.Context, scope platformtx.TenantScope, store Store,
	operation catalogdomain.Operation, authority PurgeAuthority,
) (time.Time, bool, error) {
	steps, err := application.purgeOwnerSteps(operation)
	if err != nil {
		return time.Time{}, false, err
	}
	existing, err := CompileOwnerReceipts(ctx, store, mustUUID(operation.ID.String()))
	if err != nil {
		return time.Time{}, false, err
	}
	retryAt, waiting, blocked := time.Time{}, false, false
	for _, step := range steps {
		if hasOwnerReceipt(existing, operation.ID, step.pin, "purge") {
			continue
		}
		command := purgeOwnerCommand(operation, authority, step.pin, "finalize")
		result, callErr := step.owner.FinalizeComponentPurge(ctx, scope, command)
		if callErr != nil {
			err = errors.Join(err, callErr)
		} else if !result.ValidFinalizeFor(command) || !ownerResultExact(step.pin, result) {
			err = errors.Join(err, ErrOwnerReceiptStale)
		} else if result.State == catalogcomponentv2.StatePurged {
			err = errors.Join(err, recordPurgeOwnerResult(ctx, store, operation, step.pin, result))
		} else {
			retryAt, waiting, blocked = collectPurgeState(
				retryAt, waiting, blocked, result,
			)
		}
	}
	return purgeCoordinationResult(retryAt, waiting, blocked, err)
}

func collectPurgeState(
	retryAt time.Time, waiting, blocked bool, result componentowner.OwnerResult,
) (time.Time, bool, bool) {
	switch result.State {
	case catalogcomponentv2.StateBlocked:
		blocked = true
	case catalogcomponentv2.StateWaiting, catalogcomponentv2.StateCooling:
		waiting = true
		retryAt = earlierRetry(retryAt, result.RetryAt)
	case catalogcomponentv2.StatePurgePending:
		if !result.Ready {
			waiting = true
			retryAt = earlierRetry(retryAt, result.RetryAt)
		}
	}
	return retryAt, waiting, blocked
}

func purgeCoordinationResult(
	retryAt time.Time, waiting, blocked bool, err error,
) (time.Time, bool, error) {
	if err != nil {
		return time.Time{}, false, err
	}
	if blocked {
		return time.Time{}, false, ErrDependencyBlocked
	}
	return retryAt, !waiting, nil
}

func ownerResultExact(pin catalogdomain.ComponentPin, result componentowner.OwnerResult) bool {
	return true
}

func (application *Application) purgeOwnerSteps(
	operation catalogdomain.Operation,
) ([]purgeOwnerStep, error) {
	steps := make([]purgeOwnerStep, 0, len(operation.Plan.Components))
	for _, pin := range operation.Plan.Components {
		owner, err := application.owners.Resolve(pin.Owner)
		lifecycleOwner, ok := owner.(componentowner.LifecycleOwner)
		if err != nil || owner == nil || !ok || lifecycleOwner == nil {
			return nil, ErrOwnerNotV2
		}
		steps = append(steps, purgeOwnerStep{pin: pin, owner: lifecycleOwner})
	}
	return steps, nil
}

func purgeOwnerCommand(
	operation catalogdomain.Operation, authority PurgeAuthority,
	pin catalogdomain.ComponentPin, phase string,
) componentowner.PurgeCommand {
	return componentowner.PurgeCommand{
		Authority: authority,
		Component: componentowner.Component{
			ID: pin.ComponentID.String(), Owner: pin.Owner,
			Generation: pin.OwnerGeneration, Contract: pin.Contract,
			ContentSHA256: pin.ContentSHA256,
		},
		IdempotencyKey: fmt.Sprintf(
			"%s:%s:%s", operation.ID.String(), pin.ComponentID.String(), phase,
		),
	}
}

func mustUUID(raw string) uuid.UUID {
	value, _ := uuid.Parse(raw)
	return value
}
