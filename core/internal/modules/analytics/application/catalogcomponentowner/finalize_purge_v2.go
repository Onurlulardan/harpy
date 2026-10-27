package catalogcomponentowner

import (
	"context"

	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (owner *Owner) FinalizeComponentPurge(
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
	authority, err := store.LoadPurgeAuthority(
		ctx,
		command.Component,
		packageID,
		uuid.MustParse(command.Authority.OperationID),
	)
	binding := ResourceBinding{
		ResourceID: authority.ResourceID,
		VersionID:  authority.VersionID,
		HeadID:     authority.HeadID,
	}
	if err != nil || !authorityMatches(command, binding, authority) {
		return componentowner.OwnerResult{}, unavailableOwnerResult()
	}
	receipt, err := owner.Purge(
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
	if err != nil {
		return componentowner.OwnerResult{}, err
	}
	result := purgeResult(
		command,
		uuid.MustParse(receipt.ID),
		binding,
	)
	result.AuthoritySHA256 = authority.AuthoritySHA256
	result.State = catalogcomponentv2.StatePurged
	result.Ready = true
	result.CompletedAt = receipt.CompletedAt
	result.ResultSHA256 = catalogcomponentv2.ResultSHA256(result)
	if !result.ValidFinalizeFor(command) {
		return componentowner.OwnerResult{}, unavailableOwnerResult()
	}
	return result, nil
}
