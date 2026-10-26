package catalogcomponentowner

import (
	"context"
	"fmt"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (owner *Owner) Provision(
	ctx context.Context,
	scope platformtx.TenantScope,
	request contract.ProvisionRequest,
) (contract.ProvisionReceipt, error) {
	receipt, err := owner.BaseOwner.Provision(ctx, scope, request)
	if err != nil || !receipt.Ready {
		return receipt, err
	}
	_, endpoint, err := owner.BaseOwner.FormsDependencySlugs(request)
	if err != nil {
		return contract.ProvisionReceipt{}, err
	}
	if endpoint {
		return contract.ProvisionReceipt{}, fmt.Errorf(
			"analytics forms dependencies are unavailable",
		)
	}
	return receipt, nil
}
