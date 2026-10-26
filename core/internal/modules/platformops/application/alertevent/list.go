package alertevent

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type Backend interface {
	AlertEventListing(context.Context, string, platformopsapi.AlertEventFilter) (platformopsapi.AlertEventPage, error)
	AlertEventLookup(context.Context, string, string) (platformopsapi.AlertEvent, error)
}

type AlertEventOperations struct{ backend Backend }

func NewAlertEventOperations(backend Backend) *AlertEventOperations {
	return &AlertEventOperations{backend: backend}
}

func (operations *AlertEventOperations) ListAlertEvents(
	ctx context.Context, environmentID string, filter platformopsapi.AlertEventFilter,
) (platformopsapi.AlertEventPage, error) {
	return operations.backend.AlertEventListing(ctx, environmentID, filter)
}
