package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (service *Service) Trend(ctx context.Context, access Access,
	request TimeRangeRequest, granularity domain.Granularity,
) ([]domain.TrendPoint, error) {
	if err := service.ready(ctx); err != nil {
		return nil, err
	}
	if err := authorize(access, OperationTenantTrend, OperationPlatformTrend); err != nil {
		return nil, err
	}
	window, err := service.timeRange(request)
	if err != nil || !domain.ValidGranularity(granularity) {
		return nil, domain.ErrInvalid
	}
	return service.store.Trend(ctx, access, window, granularity)
}
