package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) AlertRuleDeletion(
	ctx context.Context,
	environmentID, id string,
) error {
	if service == nil || service.alerts == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(environmentID) || !validRecoveryUUID(id) {
		return platformopsapi.ErrInvalid
	}
	if err := service.alerts.DeleteAlertRule(ctx, environmentID, id); err != nil {
		return mapRecoveryStoreError(err)
	}
	return nil
}
