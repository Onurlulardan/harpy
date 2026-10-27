package tenantmigration

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (service *Service) Enqueue(
	ctx context.Context, operationID uuid.UUID,
) (tenantapi.TenantMigrationJob, error) {
	if service == nil || operationID == uuid.Nil || ctx == nil || ctx.Err() != nil {
		return tenantapi.TenantMigrationJob{}, ErrInvalidRequest
	}
	ready, found, err := service.repository.CurrentReady(ctx)
	if err != nil {
		return tenantapi.TenantMigrationJob{}, err
	}
	if !found || !ready.IsValid() {
		return tenantapi.TenantMigrationJob{}, ErrExecutorNotReady
	}
	releaseID, err := uuid.Parse(ready.ReleaseID)
	if err != nil || releaseID == uuid.Nil {
		return tenantapi.TenantMigrationJob{}, ErrUnavailable
	}
	jobID, err := service.newID()
	if err != nil || jobID == uuid.Nil {
		return tenantapi.TenantMigrationJob{}, ErrUnavailable
	}
	job, err := service.repository.Enqueue(ctx, EnqueueBinding{
		JobID: jobID, OperationID: operationID, ReleaseID: releaseID,
		RuntimeGeneration:       ready.RuntimeGeneration,
		DatabaseGeneration:      ready.DatabaseGeneration,
		MigrationManifestSHA256: ready.MigrationManifestSHA256,
	})
	if err != nil {
		return tenantapi.TenantMigrationJob{}, err
	}
	if !job.IsValid() {
		return tenantapi.TenantMigrationJob{}, ErrUnavailable
	}
	return job, nil
}
