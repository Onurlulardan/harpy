package platformiam

import (
	"context"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *OwnerTransferService) RecordOwnerTransferAck(
	ctx context.Context,
	scope platformtx.TenantScope,
	command platformdomain.OwnerTransferAcknowledge,
) (platformdomain.OwnerTransferPrepared, error) {
	if service == nil || command.Validate(service.clock.Now()) != nil ||
		!commandScope(ctx, scope, command.ActorID) {
		return platformdomain.OwnerTransferPrepared{}, platformdomain.ErrInvalidCommand
	}
	key, found := service.trust.ResolveRegistryKey(command.Acknowledgement.RegistryKID)
	if !found || command.Acknowledgement.Verify(service.clock.Now(), key) != nil {
		return platformdomain.OwnerTransferPrepared{}, platformdomain.ErrDenied
	}
	result, err := service.store.RecordOwnerTransferAck(ctx, scope, command)
	if err != nil {
		return platformdomain.OwnerTransferPrepared{}, err
	}
	if result.Validate(command.OperationID, "ack_verified") != nil {
		return platformdomain.OwnerTransferPrepared{}, platformdomain.ErrStorage
	}
	return result, nil
}
