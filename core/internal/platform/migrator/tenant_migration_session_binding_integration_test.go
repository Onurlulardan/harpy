//go:build integration

package migrator

import (
	"context"
	"strings"
	"testing"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	"github.com/jackc/pgx/v5"
)

var tenantSessionBindingTarget = postgresstore.TenantMigrationTarget{
	TenantID:   "17171717-0000-4000-8000-000000000017",
	SchemaName: "tenant_session_binding",
}

func TestTenantMigrationSessionPostgreSQL184RequiresExistingTargetSchema(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := newTenantMigrationFixture(t, ctx)
	runner := fixture.runner(t, postgresv1.DatabaseOriginFresh)
	insertTenantTarget(t, ctx, fixture.database, tenantSessionBindingTarget)
	qualifiedSchema := pgx.Identifier{tenantSessionBindingTarget.SchemaName}.Sanitize()
	if _, err := fixture.database.ExecContext(ctx, "DROP SCHEMA "+qualifiedSchema); err != nil {
		t.Fatal(err)
	}

	_, err := runner.executeTenantTarget(ctx, tenantSessionBindingTarget, tenantRunV1)
	if err == nil || !strings.Contains(err.Error(), "target schema is absent") {
		t.Fatalf("missing target schema error = %v", err)
	}
	var checkpointCount int
	if err := fixture.database.QueryRowContext(ctx, `SELECT pg_catalog.count(*)
FROM platform.tenant_migration_checkpoints WHERE tenant_id=$1`,
		tenantSessionBindingTarget.TenantID).Scan(&checkpointCount); err != nil {
		t.Fatal(err)
	}
	if checkpointCount != 0 {
		t.Fatalf("missing target schema checkpoints = %d, want 0", checkpointCount)
	}

	if _, err := fixture.database.ExecContext(ctx,
		"CREATE SCHEMA "+qualifiedSchema+" AUTHORIZATION harpy_owner"); err != nil {
		t.Fatal(err)
	}
	if _, err := runner.executeTenantTarget(
		ctx, tenantSessionBindingTarget, tenantRunV1,
	); err != nil {
		t.Fatalf("bind existing target schema: %v", err)
	}
	assertTenantMigrationApplied(
		t, ctx, fixture.database, tenantSessionBindingTarget, "fresh",
	)
}
