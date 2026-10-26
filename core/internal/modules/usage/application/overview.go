package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (service *Service) Overview(ctx context.Context, access Access,
	request TimeRangeRequest,
) (domain.Overview, error) {
	if err := service.ready(ctx); err != nil {
		return domain.Overview{}, err
	}
	if err := authorize(access, OperationTenantOverview, OperationPlatformOverview); err != nil {
		return domain.Overview{}, err
	}
	window, err := service.timeRange(request)
	if err != nil {
		return domain.Overview{}, err
	}
	value, err := service.store.Overview(ctx, access, window, service.clock.Now().UTC())
	return decorateOverview(value), err
}
