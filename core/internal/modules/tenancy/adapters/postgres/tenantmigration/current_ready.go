package tenantmigration

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	tenantapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantmigration"
)

func (storage *repository) CurrentReady(
	ctx context.Context,
) (tenantapi.TenantMigrationReady, bool, error) {
	if !storage.usable(ctx) {
		return tenantapi.TenantMigrationReady{}, false, tenantapp.ErrUnavailable
	}
	var ready tenantapi.TenantMigrationReady
	var found bool
	err := storage.queryRow(ctx, currentReadySQL).Scan(
		&ready.ReleaseID, &ready.EnvironmentID, &ready.DeploymentMode,
		&ready.RuntimeGeneration, &ready.DatabaseGeneration,
		&ready.MigrationManifestSHA256, &ready.ReadyUntil, &found,
	)
	if err != nil || !found {
		return tenantapi.TenantMigrationReady{}, found, err
	}
	return ready, true, nil
}
