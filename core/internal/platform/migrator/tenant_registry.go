package migrator

import (
	"context"
	"database/sql"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func tenantInfrastructureRequired(
	ctx context.Context,
	connection *sql.Conn,
) (bool, error) {
	return postgresstore.TenantInfrastructureExists(ctx, connection)
}

func listTenantTargets(
	ctx context.Context,
	connection *sql.Conn,
) ([]postgresstore.TenantMigrationTarget, error) {
	return postgresstore.ListTenantMigrationTargets(ctx, connection)
}
