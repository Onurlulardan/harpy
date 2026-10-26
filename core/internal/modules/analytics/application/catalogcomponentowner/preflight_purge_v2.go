package catalogcomponentowner

import (
	"context"

	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (owner *Owner) PreflightComponentPurge(
	ctx context.Context,
	scope platformtx.TenantScope,
	command componentowner.PurgeCommand,
) (componentowner.OwnerResult, error) {
	store, packageID, operationID, componentID, err :=
		owner.purgeStore(ctx, scope, command)
	if err != nil {
		return componentowner.OwnerResult{}, err
	}
	snapshot, err := store.Preflight(
		ctx,
		command.Component,
		packageID,
		command.Authority.ReferenceEpoch,
	)
	if err != nil {
		return componentowner.OwnerResult{}, err
	}
	result := purgeResult(
		command,
		deriveReceiptID(operationID, componentID, "preflight"),
		snapshot.ResourceBinding,
	)
	if len(snapshot.Blockers) == 0 {
		result.State = catalogcomponentv2.StateReady
		result.Ready = true
		result.CompletedAt = command.Authority.RequestedAt
	} else {
		result.State = catalogcomponentv2.StateBlocked
		result.Blockers = append(
			[]componentowner.Blocker(nil),
			snapshot.Blockers...,
		)
	}
	result.ResultSHA256 = catalogcomponentv2.ResultSHA256(result)
	if !result.ValidPreflightFor(command) {
		return componentowner.OwnerResult{}, unavailableOwnerResult()
	}
	return result, nil
}
