package recoveryanchors

import (
	"context"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *Service) DeleteRecoveryAnchor(
	ctx context.Context, command platformopsapi.RecoveryAnchorDeleteCommand,
) (platformopsapi.RecoveryAnchorOperation, error) {
	if service == nil || service.disabled || service.anchors == nil {
		return platformopsapi.RecoveryAnchorOperation{}, platformopsapi.ErrUnavailable
	}
	result, err := service.anchors.Delete(ctx, anchorapi.DeleteCommand{
		Subject:     iamSubject(command.RecoveryAnchorSubject),
		OperationID: command.OperationID, AnchorID: command.AnchorID,
		IdempotencyKey: command.IdempotencyKey, RequestSHA256: command.RequestSHA256})
	if err != nil {
		return platformopsapi.RecoveryAnchorOperation{}, mapServiceError(err)
	}
	return platformopsapi.RecoveryAnchorOperation{OperationID: result.OperationID,
		AnchorID: result.AnchorID, Kind: result.Kind, State: result.State}, nil
}

var _ platformopsapi.RecoveryAnchors = (*Service)(nil)
