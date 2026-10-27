package catalogcomponentowner

import (
	"context"
	"reflect"

	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (owner *Owner) BeginComponentPurge(
	ctx context.Context,
	scope platformtx.TenantScope,
	command componentowner.PurgeCommand,
) (componentowner.OwnerResult, error) {
	store, packageID, _, _, err := owner.purgeStore(
		ctx,
		scope,
		command,
	)
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
	receipt, err := owner.BaseOwner.PurgeBlockers(
		ctx,
		scope,
		componentowner.Command{
			OperationID:    command.Authority.OperationID,
			Component:      command.Component,
			Action:         "purge",
			ReferenceEpoch: command.Authority.ReferenceEpoch,
			IdempotencyKey: command.IdempotencyKey,
		},
	)
	if err != nil || receipt.OperationID != command.Authority.OperationID ||
		receipt.ComponentID != command.Component.ID ||
		receipt.ReferenceEpoch != command.Authority.ReferenceEpoch ||
		!reflect.DeepEqual(receipt.Blockers, snapshot.Blockers) {
		return componentowner.OwnerResult{}, unavailableOwnerResult()
	}
	authority, err := store.PersistPurgeAuthority(
		ctx,
		command.Component,
		packageID,
		PurgeAuthorityBinding{
			OperationID:     uuid.MustParse(command.Authority.OperationID),
			ComponentID:     uuid.MustParse(command.Component.ID),
			ReferenceEpoch:  int64(command.Authority.ReferenceEpoch),
			AuthoritySHA256: command.Authority.AuthoritySHA256,
			ResourceID:      snapshot.ResourceID,
			VersionID:       snapshot.VersionID,
			HeadID:          snapshot.HeadID,
			RecordedAt:      receipt.CompletedAt,
		},
	)
	if err != nil ||
		!authorityMatches(command, snapshot.ResourceBinding, authority) {
		return componentowner.OwnerResult{}, unavailableOwnerResult()
	}
	result := purgeResult(
		command,
		uuid.MustParse(receipt.ID),
		snapshot.ResourceBinding,
	)
	result.AuthoritySHA256 = authority.AuthoritySHA256
	if len(receipt.Blockers) == 0 {
		result.State = catalogcomponentv2.StatePurgePending
		result.Ready = true
		result.CompletedAt = receipt.CompletedAt
	} else {
		result.State = catalogcomponentv2.StateBlocked
		result.Blockers = append(
			[]componentowner.Blocker(nil),
			receipt.Blockers...,
		)
	}
	result.ResultSHA256 = catalogcomponentv2.ResultSHA256(result)
	if !result.ValidBeginFor(command) {
		return componentowner.OwnerResult{}, unavailableOwnerResult()
	}
	return result, nil
}
