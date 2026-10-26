package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) AlertTestRecording(
	ctx context.Context,
	environmentID, id, actorID string,
) (platformopsapi.AlertEvent, error) {
	if service == nil || service.alerts == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(environmentID) || !validRecoveryUUID(id) ||
		!validRecoveryUUID(actorID) {
		return platformopsapi.AlertEvent{}, platformopsapi.ErrInvalid
	}
	result, err := service.alerts.RecordAlertTest(ctx, environmentID, id, actorID,
		"Durable alert test event recorded by an authorized platform actor")
	if err != nil {
		return platformopsapi.AlertEvent{}, mapRecoveryStoreError(err)
	}
	return result, nil
}
