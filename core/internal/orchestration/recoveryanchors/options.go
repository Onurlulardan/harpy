package recoveryanchors

import (
	"context"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *Service) CreateRecoveryAnchorOptions(
	ctx context.Context, command platformopsapi.RecoveryAnchorOptionsCommand,
) (platformopsapi.RecoveryAnchorRegistrationOptions, error) {
	if service == nil || service.disabled || service.anchors == nil {
		return platformopsapi.RecoveryAnchorRegistrationOptions{}, platformopsapi.ErrUnavailable
	}
	result, err := service.anchors.CreateOptions(ctx, anchorapi.OptionsCommand{
		Subject:     iamSubject(command.RecoveryAnchorSubject),
		OperationID: command.OperationID, AnchorID: command.AnchorID,
		IdempotencyKey: command.IdempotencyKey, RequestSHA256: command.RequestSHA256})
	if err != nil {
		return platformopsapi.RecoveryAnchorRegistrationOptions{}, mapServiceError(err)
	}
	return platformopsapi.RecoveryAnchorRegistrationOptions{
		OperationID: result.OperationID, Challenge: result.Challenge,
		RPID: result.RPID, UserID: result.UserID,
		ExpiresAt: result.ExpiresAt, Replayed: result.Replayed}, nil
}
