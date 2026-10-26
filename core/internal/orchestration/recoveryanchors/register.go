package recoveryanchors

import (
	"context"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *Service) RegisterRecoveryAnchor(
	ctx context.Context, command platformopsapi.RecoveryAnchorRegisterCommand,
) (platformopsapi.RecoveryAnchorOperation, error) {
	if service == nil || service.disabled || service.anchors == nil {
		return platformopsapi.RecoveryAnchorOperation{}, platformopsapi.ErrUnavailable
	}
	result, err := service.anchors.Register(ctx, anchorapi.RegisterCommand{
		Subject:     iamSubject(command.RecoveryAnchorSubject),
		OperationID: command.OperationID, IdempotencyKey: command.IdempotencyKey,
		RequestSHA256:  command.RequestSHA256,
		CredentialJSON: append([]byte(nil), command.CredentialJSON...)})
	if err != nil {
		return platformopsapi.RecoveryAnchorOperation{}, mapServiceError(err)
	}
	return platformopsapi.RecoveryAnchorOperation{OperationID: result.OperationID,
		AnchorID: result.AnchorID, Kind: result.Kind, State: result.State}, nil
}
