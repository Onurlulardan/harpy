package alertevent

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (operations *AlertEventOperations) GetAlertEvent(
	ctx context.Context, environmentID, eventID string,
) (platformopsapi.AlertEvent, error) {
	return operations.backend.AlertEventLookup(ctx, environmentID, eventID)
}
