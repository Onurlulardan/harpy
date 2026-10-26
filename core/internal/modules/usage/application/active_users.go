package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (service *Service) ActiveUsers(ctx context.Context, access Access) (domain.ActiveUsers, error) {
	if err := service.ready(ctx); err != nil {
		return domain.ActiveUsers{}, err
	}
	if err := authorize(access, OperationTenantActiveUsers, OperationPlatformActiveUsers); err != nil {
		return domain.ActiveUsers{}, err
	}
	return service.store.ActiveUsers(ctx, access, service.clock.Now().UTC())
}
