package tenantmigration

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (service *Service) CurrentReady(
	ctx context.Context,
) (tenantapi.TenantMigrationReady, bool, error) {
	if service == nil || service.repository == nil || ctx == nil || ctx.Err() != nil {
		return tenantapi.TenantMigrationReady{}, false, ErrUnavailable
	}
	ready, found, err := service.repository.CurrentReady(ctx)
	if err != nil || !found {
		return tenantapi.TenantMigrationReady{}, found, err
	}
	if !ready.IsValid() {
		return tenantapi.TenantMigrationReady{}, false, ErrUnavailable
	}
	return ready, true, nil
}
