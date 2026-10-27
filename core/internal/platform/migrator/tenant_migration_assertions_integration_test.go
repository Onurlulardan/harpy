//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"testing"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func assertTenantMigrationApplied(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	target postgresstore.TenantMigrationTarget,
	origin string,
) {
	t.Helper()
	var relationSchema string
	err := database.QueryRowContext(ctx, `SELECT namespace.nspname
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
WHERE relation.relname='tenancy_schema_identity' AND namespace.nspname=$1`,
		target.SchemaName).Scan(&relationSchema)
	if err != nil || relationSchema != target.SchemaName {
		t.Fatalf("tenant relation namespace = %q err=%v", relationSchema, err)
	}
	query := "SELECT tenant_id::text,schema_name,origin FROM " +
		quoteIntegrationIdentifier(target.SchemaName) + ".tenancy_schema_identity"
	var tenantID, schemaName, storedOrigin string
	if err := database.QueryRowContext(ctx, query).Scan(
		&tenantID, &schemaName, &storedOrigin,
	); err != nil || tenantID != target.TenantID || schemaName != target.SchemaName ||
		storedOrigin != origin {
		t.Fatalf("tenant sentinel = %s %s %s err=%v", tenantID, schemaName, storedOrigin, err)
	}
	var succeeded int
	if err := database.QueryRowContext(ctx, `SELECT count(*)
FROM platform.tenant_migration_checkpoints
WHERE tenant_id=$1 AND state='succeeded'`, target.TenantID).Scan(&succeeded); err != nil ||
		succeeded != 4 {
		t.Fatalf("tenant checkpoints = %d err=%v", succeeded, err)
	}
}

func assertTenantIsolation(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var crossA, crossB int
	queryA := "SELECT count(*) FROM " + quoteIntegrationIdentifier(tenantTargetA.SchemaName) +
		".tenancy_schema_identity WHERE tenant_id=$1"
	queryB := "SELECT count(*) FROM " + quoteIntegrationIdentifier(tenantTargetB.SchemaName) +
		".tenancy_schema_identity WHERE tenant_id=$1"
	if err := database.QueryRowContext(ctx, queryA, tenantTargetB.TenantID).Scan(&crossA); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRowContext(ctx, queryB, tenantTargetA.TenantID).Scan(&crossB); err != nil {
		t.Fatal(err)
	}
	if crossA != 0 || crossB != 0 {
		t.Fatalf("tenant sentinel crossed schemas: %d %d", crossA, crossB)
	}
}

func quoteIntegrationIdentifier(value string) string {
	return `"` + value + `"`
}
