//go:build integration

package migrator

import (
	"context"
	"testing"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	"github.com/jackc/pgx/v5"
)

var tenantTargetRetry = postgresstore.TenantMigrationTarget{
	TenantID: "dddddddd-0000-4000-8000-000000000004", SchemaName: "tenant_retry",
}

func TestTenantMigrationExecutorPostgreSQL184CrashRetryAndV1Sentinel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := newTenantMigrationFixture(t, ctx)
	runner := fixture.runner(t, postgresv1.DatabaseOriginFresh)
	insertTenantTarget(t, ctx, fixture.database, tenantTargetRetry)
	database := openOneShotTenantDatabase(fixture.config.Copy())
	if err := postgresstore.AcquireTenantMigrationLock(
		ctx, database, tenantTargetRetry.TenantID,
	); err != nil {
		t.Fatal(err)
	}
	searchPath := pgx.Identifier{tenantTargetRetry.SchemaName}.Sanitize() + ",pg_catalog"
	if err := postgresstore.SetTenantMigrationSearchPath(ctx, database, searchPath); err != nil {
		t.Fatal(err)
	}
	session, err := postgresstore.ReadTenantMigrationSession(ctx, database)
	if err != nil {
		t.Fatal(err)
	}
	started, err := prepareTenantCheckpoint(
		ctx, database, tenantTargetRetry, session, runner.tenantSelected[0],
		runner.tenantAliases, tenantRunOne,
	)
	if err != nil || !started {
		t.Fatalf("prepare crashed attempt: started=%v err=%v", started, err)
	}
	if err := postgresstore.ReleaseTenantMigrationLock(
		ctx, database, tenantTargetRetry.TenantID,
	); err != nil {
		t.Fatal(err)
	}
	_ = database.Close()
	if _, err := runner.executeTenantTarget(ctx, tenantTargetRetry, tenantRunTwo); err != nil {
		t.Fatal(err)
	}
	var attempt int
	var state string
	if err := fixture.database.QueryRowContext(ctx, `SELECT attempt,state
FROM platform.tenant_migration_checkpoints
WHERE tenant_id=$1 AND logical_version=$2`, tenantTargetRetry.TenantID,
		runner.tenantSelected[0].Version).Scan(&attempt, &state); err != nil ||
		attempt != 2 || state != "succeeded" {
		t.Fatalf("retry checkpoint = attempt %d state %s err=%v", attempt, state, err)
	}
	var failed, succeeded int
	if err := fixture.database.QueryRowContext(ctx, `SELECT
count(*) FILTER (WHERE state='failed' AND error_code='migration_interrupted'),
count(*) FILTER (WHERE state='succeeded')
FROM platform.tenant_migration_checkpoint_attempts
WHERE tenant_id=$1 AND logical_version=$2`, tenantTargetRetry.TenantID,
		runner.tenantSelected[0].Version).Scan(&failed, &succeeded); err != nil ||
		failed != 1 || succeeded != 1 {
		t.Fatalf("retry receipts = failed %d succeeded %d err=%v", failed, succeeded, err)
	}
	v1Target := postgresstore.TenantMigrationTarget{
		TenantID: "eeeeeeee-0000-4000-8000-000000000005", SchemaName: "tenant_v1",
	}
	insertTenantTarget(t, ctx, fixture.database, v1Target)
	v1Runner := fixture.runner(t, postgresv1.DatabaseOriginFresh)
	if _, err := v1Runner.executeTenantTarget(ctx, v1Target, tenantRunV1); err != nil {
		t.Fatal(err)
	}
	assertTenantMigrationApplied(t, ctx, fixture.database, v1Target, "fresh")
}
