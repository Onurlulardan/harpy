//go:build integration

package transaction

import (
	"context"
	"testing"
)

func seedScopeEnvironmentIdentity(
	t *testing.T,
	ctx context.Context,
	database scopeTestDatabase,
) {
	t.Helper()
	adminExecute(t, ctx, database, `INSERT INTO platform.environment_identity
SELECT true,$1,'fresh',system_identifier::text,database.oid,$2
FROM pg_catalog.pg_control_system(),pg_catalog.pg_database AS database
WHERE database.datname=pg_catalog.current_database()`, scopeEnvironmentID, scopeReleaseID)
}
