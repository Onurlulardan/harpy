package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (service *Service) Heatmap(ctx context.Context, access Access,
	request TimeRangeRequest,
) ([]domain.HeatmapCell, error) {
	if err := service.ready(ctx); err != nil {
		return nil, err
	}
	if err := authorize(access, OperationTenantHeatmap, OperationPlatformHeatmap); err != nil {
		return nil, err
	}
	window, err := service.timeRange(request)
	if err != nil {
		return nil, err
	}
	return service.store.Heatmap(ctx, access, window)
}
