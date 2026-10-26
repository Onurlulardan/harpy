package platformiam

import (
	"context"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *OwnerTransferService) CompleteOwnerTransfer(
	ctx context.Context,
	scope platformtx.TenantScope,
	command platformdomain.OwnerTransferCommit,
) (platformdomain.OwnerTransferCommitted, error) {
	if service == nil || command.Validate() != nil ||
		!commandScope(ctx, scope, command.ActorID) {
		return platformdomain.OwnerTransferCommitted{}, platformdomain.ErrInvalidCommand
	}
	result, err := service.store.CompleteOwnerTransfer(ctx, scope, command)
	if err != nil {
		return platformdomain.OwnerTransferCommitted{}, err
	}
	if result.Validate(command.OperationID) != nil {
		return platformdomain.OwnerTransferCommitted{}, platformdomain.ErrStorage
	}
	return result, nil
}
