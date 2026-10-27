package tenantmigration

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	tenantapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantmigration"
	"github.com/google/uuid"
)

func (storage *repository) Get(
	ctx context.Context, operationID uuid.UUID,
) (tenantapi.TenantMigrationJob, bool, error) {
	if !storage.usable(ctx) {
		return tenantapi.TenantMigrationJob{}, false, tenantapp.ErrUnavailable
	}
	job, found, err := scanJob(storage.queryRow(ctx, getSQL, operationID), true)
	return job, found, err
}
