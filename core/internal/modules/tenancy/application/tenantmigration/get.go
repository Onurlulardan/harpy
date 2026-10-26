package tenantmigration

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (service *Service) Get(
	ctx context.Context, operationID uuid.UUID,
) (tenantapi.TenantMigrationJob, bool, error) {
	if service == nil || operationID == uuid.Nil || ctx == nil || ctx.Err() != nil {
		return tenantapi.TenantMigrationJob{}, false, ErrInvalidRequest
	}
	job, found, err := service.repository.Get(ctx, operationID)
	if err != nil || !found {
		return tenantapi.TenantMigrationJob{}, found, err
	}
	if !job.IsValid() {
		return tenantapi.TenantMigrationJob{}, false, ErrUnavailable
	}
	return job, true, nil
}
