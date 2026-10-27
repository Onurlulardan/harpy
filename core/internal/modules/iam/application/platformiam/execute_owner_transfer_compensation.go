package platformiam

import (
	"context"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *OwnerTransferService) ExecuteOwnerTransferCompensation(
	ctx context.Context,
	scope platformtx.TenantScope,
	command platformdomain.OwnerTransferAcknowledge,
) (platformdomain.OwnerTransferCompensated, error) {
	if service == nil || command.Validate(service.clock.Now()) != nil ||
		!commandScope(ctx, scope, command.ActorID) {
		return platformdomain.OwnerTransferCompensated{}, platformdomain.ErrInvalidCommand
	}
	key, found := service.trust.ResolveRegistryKey(command.Acknowledgement.RegistryKID)
	if !found || command.Acknowledgement.Verify(service.clock.Now(), key) != nil {
		return platformdomain.OwnerTransferCompensated{}, platformdomain.ErrDenied
	}
	result, err := service.store.ExecuteOwnerTransferCompensation(ctx, scope, command)
	if err != nil {
		return platformdomain.OwnerTransferCompensated{}, err
	}
	if result.Validate(command.OperationID, result.OriginalOperationID) != nil {
		return platformdomain.OwnerTransferCompensated{}, platformdomain.ErrStorage
	}
	return result, nil
}
