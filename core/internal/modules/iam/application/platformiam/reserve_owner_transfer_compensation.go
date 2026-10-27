package platformiam

import (
	"context"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *OwnerTransferService) ReserveOwnerTransferCompensation(
	ctx context.Context,
	scope platformtx.TenantScope,
	command platformdomain.OwnerTransferCompensation,
) (platformdomain.OwnerTransferPrepared, error) {
	if service == nil || command.Validate(service.clock.Now()) != nil ||
		!commandScope(ctx, scope, command.ActorID) {
		return platformdomain.OwnerTransferPrepared{}, platformdomain.ErrInvalidCommand
	}
	result, err := service.store.ReserveOwnerTransferCompensation(ctx, scope, command)
	if err != nil {
		return platformdomain.OwnerTransferPrepared{}, err
	}
	if result.Validate(command.OperationID, "prepared") != nil {
		return platformdomain.OwnerTransferPrepared{}, platformdomain.ErrStorage
	}
	return result, nil
}
