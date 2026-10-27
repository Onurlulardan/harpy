package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (service *Service) UserSummary(ctx context.Context, access Access, subject string,
	request TimeRangeRequest,
) (domain.UserSummary, error) {
	if err := service.ready(ctx); err != nil {
		return domain.UserSummary{}, err
	}
	if err := authorize(access, OperationTenantUserSummary, OperationPlatformUserSummary); err != nil {
		return domain.UserSummary{}, err
	}
	lookup, err := domain.NewTimelineQuery(access.Scope(), subject, 1, 1)
	if err != nil {
		return domain.UserSummary{}, err
	}
	window, err := service.timeRange(request)
	if err != nil {
		return domain.UserSummary{}, err
	}
	value, err := service.store.UserSummary(ctx, access, lookup.Subject, window)
	value.TopRoutes = decorateRoutes(value.TopRoutes)
	return value, err
}
