package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) SystemHealthRead(
	ctx context.Context,
) (platformopsapi.SystemHealth, error) {
	if service == nil || service.health == nil || ctx == nil || ctx.Err() != nil {
		return platformopsapi.SystemHealth{}, platformopsapi.ErrUnavailable
	}
	return service.health.ReadSystemHealth(ctx)
}
