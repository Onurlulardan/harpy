package recoveryanchors

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *Service) ListRecoveryAnchors(
	ctx context.Context, subject platformopsapi.RecoveryAnchorSubject,
) ([]platformopsapi.RecoveryAnchor, error) {
	if service == nil || service.disabled || service.anchors == nil {
		return nil, platformopsapi.ErrUnavailable
	}
	stored, err := service.anchors.List(ctx, iamSubject(subject))
	if err != nil {
		return nil, mapServiceError(err)
	}
	result := make([]platformopsapi.RecoveryAnchor, len(stored))
	for index := range stored {
		result[index] = platformopsapi.RecoveryAnchor{AnchorID: stored[index].AnchorID,
			Generation:   stored[index].Generation,
			CredentialID: append([]byte(nil), stored[index].CredentialID...),
			RPID:         stored[index].RPID, UserVerification: stored[index].UserVerification,
			BackupEligible: stored[index].BackupEligible,
			BackupState:    stored[index].BackupState, State: stored[index].State}
	}
	return result, nil
}
