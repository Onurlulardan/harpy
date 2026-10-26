package tenantmigration

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	tenantapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantmigration"
)

func (storage *repository) Enqueue(
	ctx context.Context, binding tenantapp.EnqueueBinding,
) (tenantapi.TenantMigrationJob, error) {
	if !storage.usable(ctx) {
		return tenantapi.TenantMigrationJob{}, tenantapp.ErrUnavailable
	}
	job, err := scanEnqueue(storage.queryRow(ctx, enqueueSQL, binding.JobID,
		binding.OperationID, binding.ReleaseID, binding.RuntimeGeneration,
		binding.DatabaseGeneration, binding.MigrationManifestSHA256))
	if err != nil {
		return tenantapi.TenantMigrationJob{}, err
	}
	current, found, err := storage.Get(ctx, binding.OperationID)
	if err != nil || !found || current.JobID != job.JobID {
		return tenantapi.TenantMigrationJob{}, tenantapp.ErrUnavailable
	}
	return current, nil
}
