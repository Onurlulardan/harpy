package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (service *Service) Features(ctx context.Context, access Access,
	request TimeRangeRequest, limit int,
) ([]domain.FeatureUsage, error) {
	if err := service.ready(ctx); err != nil {
		return nil, err
	}
	if err := authorize(access, OperationTenantFeatures, OperationPlatformFeatures); err != nil {
		return nil, err
	}
	window, err := service.timeRange(request)
	if err != nil {
		return nil, err
	}
	query, err := domain.NewRankedQuery(window, limit, 100)
	if err != nil {
		return nil, err
	}
	values, err := service.store.Features(ctx, access, query)
	return decorateFeatures(values), err
}
