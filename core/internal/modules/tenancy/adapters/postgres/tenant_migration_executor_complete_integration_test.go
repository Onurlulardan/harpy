//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func seedCompletedTenantRun(
	t *testing.T, ctx context.Context, database *sql.DB,
	claim executorClaim, origin string,
) {
	t.Helper()
	execTenantExecutorFixture(t, ctx, database, `INSERT INTO platform.migration_runs (
run_id,release_id,environment_id,database_origin,command,state,
migration_manifest_sha256,extension_manifest_sha256,roles_manifest_sha256,
privileges_manifest_sha256,ownership_manifest_sha256,started_at,finished_at)
VALUES ($1,$2,$3,$4,'tenant_migration_executor','succeeded',
$5,$6,$7,$8,$9,now(),now())`, claim.runID, tenantExecutorReleaseID,
		testEnvironmentID, origin, strings.Repeat("a", 64), strings.Repeat("b", 64),
		strings.Repeat("c", 64), strings.Repeat("d", 64), strings.Repeat("e", 64))
	execTenantExecutorFixture(t, ctx, database, `INSERT INTO
platform.tenant_migration_checkpoints (
tenant_id,owner,logical_version,database_origin,selected_path,selected_sha256,
run_id,state,attempt,started_at,finished_at,updated_at)
VALUES ($1,'tenancy',20260714011000,$2,'tenancy/new.sql',$3,$4,
'succeeded',1,now(),now(),now())`, claim.tenantID, origin,
		strings.Repeat("c", 64), claim.runID)
}

func completeTenantMigration(
	t *testing.T, ctx context.Context, database *sql.DB, claim executorClaim,
) {
	t.Helper()
	connection := tenantMigratorConnection(t, ctx, database)
	defer connection.Close()
	var catalog string
	var requirements int
	err := connection.QueryRowContext(ctx, `SELECT * FROM
platform.tenancy_tenant_migration_complete_v3($1,$2,$3,$4,$5,$6,$7)`,
		claim.jobID, claim.attempt, claim.profileID, claim.leaseID, uuid.New(),
		strings.Repeat("6", 64), strings.Repeat("7", 64)).
		Scan(&catalog, &requirements)
	if err != nil || catalog != strings.Repeat("b", 64) || requirements != 1 {
		t.Fatalf("complete tenant migration: %s/%d %v", catalog, requirements, err)
	}
}

func assertTenantRemainsInvisible(
	t *testing.T, ctx context.Context, database *sql.DB, tenantID uuid.UUID,
) {
	t.Helper()
	var count int
	if err := database.QueryRowContext(ctx,
		"SELECT count(*) FROM platform.tenants WHERE id=$1", tenantID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("tenant became visible before WP9: %d", count)
	}
}
