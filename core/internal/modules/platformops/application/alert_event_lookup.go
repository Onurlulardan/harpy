package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) AlertEventLookup(
	ctx context.Context, environmentID, eventID string,
) (platformopsapi.AlertEvent, error) {
	if service == nil || service.alerts == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(environmentID) || !validRecoveryUUID(eventID) {
		return platformopsapi.AlertEvent{}, platformopsapi.ErrInvalid
	}
	result, err := service.alerts.GetAlertEvent(ctx, environmentID, eventID)
	if err != nil {
		return platformopsapi.AlertEvent{}, mapRecoveryStoreError(err)
	}
	return result, nil
}
