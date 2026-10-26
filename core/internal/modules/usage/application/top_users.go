package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (service *Service) TopUsers(ctx context.Context, access Access,
	request TimeRangeRequest, limit int,
) ([]domain.TopUser, error) {
	if err := service.ready(ctx); err != nil {
		return nil, err
	}
	if err := authorize(access, OperationTenantTopUsers, OperationPlatformTopUsers); err != nil {
		return nil, err
	}
	window, err := service.timeRange(request)
	if err != nil {
		return nil, err
	}
	query, err := domain.NewRankedQuery(window, limit, 50)
	if err != nil {
		return nil, err
	}
	return service.store.TopUsers(ctx, access, query)
}
