package migrator

import (
	"context"

	tenantmigrationstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/tenantmigrationexecutor/adapters/postgres"
	tenantexec "github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
)

func (executor *TenantExecutor) ensureSchema(
	ctx context.Context, claim tenantexec.Claim,
) error {
	return tenantmigrationstore.EnsureSchema(
		ctx, executor.database, executor.newID, claim,
	)
}
