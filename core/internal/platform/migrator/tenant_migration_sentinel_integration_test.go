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

func TestTenantMigrationPostgreSQL184RejectsSentinelMismatchOnVerify(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := newTenantMigrationFixture(t, ctx)
	runner := fixture.runner(t, postgresv1.DatabaseOriginFresh)
	target := postgresstore.TenantMigrationTarget{
		TenantID: "15151515-0000-4000-8000-000000000011", SchemaName: "tenant_bad_sentinel",
	}
	insertTenantTarget(t, ctx, fixture.database, target)
	if _, err := runner.executeTenantTarget(ctx, target, tenantRunOne); err != nil {
		t.Fatal(err)
	}
	qualified := quoteIntegrationIdentifier(target.SchemaName) + ".tenancy_schema_identity"
	if _, err := fixture.database.ExecContext(ctx, "ALTER TABLE "+qualified+
		" DISABLE TRIGGER tenancy_schema_identity_reject_mutation"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.database.ExecContext(ctx, "UPDATE "+qualified+
		" SET origin='v1_adopted'"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.database.ExecContext(ctx, "ALTER TABLE "+qualified+
		" ENABLE TRIGGER tenancy_schema_identity_reject_mutation"); err != nil {
		t.Fatal(err)
	}
	database := openOneShotTenantDatabase(fixture.config.Copy())
	defer database.Close()
	if err := postgresstore.AcquireTenantMigrationLock(ctx, database, target.TenantID); err != nil {
		t.Fatal(err)
	}
	defer postgresstore.ReleaseTenantMigrationLock(ctx, database, target.TenantID)
	searchPath := pgx.Identifier{target.SchemaName}.Sanitize() + ",pg_catalog"
	if err := postgresstore.SetTenantMigrationSearchPath(ctx, database, searchPath); err != nil {
		t.Fatal(err)
	}
	proof, err := postgresstore.ReadTenantMigrationSession(ctx, database)
	if err != nil {
		t.Fatal(err)
	}
	err = runner.verifyTenantIdentity(ctx, database, target, proof)
	if err == nil || !strings.Contains(err.Error(), "identity is not exact") {
		t.Fatalf("sentinel mismatch verification error = %v", err)
	}
}
