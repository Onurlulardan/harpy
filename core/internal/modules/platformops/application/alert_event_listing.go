package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) AlertEventListing(
	ctx context.Context,
	environmentID string,
	filter platformopsapi.AlertEventFilter,
) (platformopsapi.AlertEventPage, error) {
	if service == nil || service.alerts == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(environmentID) || filter.Page < 1 || filter.Limit < 1 ||
		filter.Limit > 100 ||
		len(filter.ServiceName) > 100 || !validAlertEventStatus(filter.Status) {
		return platformopsapi.AlertEventPage{}, platformopsapi.ErrInvalid
	}
	result, err := service.alerts.ListAlertEvents(ctx, environmentID, filter)
	if err != nil {
		return platformopsapi.AlertEventPage{}, mapRecoveryStoreError(err)
	}
	return result, nil
}

func validAlertEventStatus(value string) bool {
	switch value {
	case "", "healthy", "degraded", "unhealthy", "unknown":
		return true
	default:
		return false
	}
}
