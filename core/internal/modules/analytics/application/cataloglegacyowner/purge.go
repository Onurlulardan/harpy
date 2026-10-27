package cataloglegacyowner

import (
	"context"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (owner *Owner) Purge(
	ctx context.Context,
	scope platformtx.TenantScope,
	command contract.Command,
) (contract.Receipt, error) {
	receipt, err := owner.BaseOwner.Purge(ctx, scope, command)
	if err != nil {
		return contract.Receipt{}, err
	}
	if err = owner.decommission(
		ctx,
		scope,
		command,
		receipt,
	); err != nil {
		return contract.Receipt{}, err
	}
	return receipt, nil
}
