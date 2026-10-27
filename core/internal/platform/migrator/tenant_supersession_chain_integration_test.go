//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	chainVersion = int64(20260730120104)
	chainRoot    = "core/internal/modules/system_resource_layout/" +
		"migrations/tenant/"
	chainBase       = "20260730120104_repair_installation_layout_publications_v1"
	chainOriginal   = chainRoot + "fresh/" + chainBase + ".fresh.sql"
	chainFirst      = chainRoot + "supersessions/fresh/" + chainBase + ".fresh.sql"
	chainSecond     = chainRoot + "supersessions/v2/fresh/" + chainBase + ".fresh.sql"
	chainTerminal   = chainRoot + "supersessions/v3/fresh/" + chainBase + ".fresh.sql"
	chainOriginalSH = "561757a212e2205e14091db2f9c05ab94f97d5a8cc67438d16c44a2ed4d0a8a2"
	chainFirstSH    = "9ecca5ceb2b0453956a81f46e46d2bf633890ee5e8c212002b33b35d73faa575"
	chainSecondSH   = "596819b7a320d41d283ea4d88846566a16bcbe6e6739701b64af77bca93d524d"
	chainTerminalSH = "cef35a91f7e9cffebc4a48a3c26d69b5edb90cbd27afae1e67dc606c58309972"
	chainOldRunID   = "018f0000-0000-7000-8000-000000000310"
	chainNewRunID   = "018f0000-0000-7000-8000-000000000320"
)

func TestTenantSupersessionChainAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	database, _ := startRunnerIntegrationDatabase(t, ctx)
	prepareTenantSupersessionChainPG18(t, ctx, database)
	assertTenantSupersessionV3Registry(t, ctx, database)
	assertCrossOwnerSupersessionEdgeRejected(t, ctx, database)

	fixtures := []struct {
		id, path, sha string
	}{
		{"018f0000-0000-7000-8000-000000000310",
			chainSecond, chainSecondSH},
		{"018f0000-0000-7000-8000-000000000311",
			chainFirst, chainFirstSH},
		{"018f0000-0000-7000-8000-000000000312",
			chainOriginal, chainOriginalSH},
	}
	for _, fixture := range fixtures {
		seedFailedChainCheckpoint(t, ctx, database, fixture.id, fixture.path, fixture.sha)
		transitionChainCheckpoint(t, ctx, database, fixture.id)
	}
	assertSucceededAliasesSatisfyTerminal(t, ctx, database)
}

func prepareTenantSupersessionChainPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `CREATE SCHEMA platform
AUTHORIZATION harpy_owner;
CREATE TABLE platform.migration_runs (run_id uuid PRIMARY KEY);`)
	if err == nil {
		_, err = database.ExecContext(ctx,
			"INSERT INTO platform.migration_runs VALUES ($1),($2)",
			chainOldRunID, chainNewRunID)
	}
	if err != nil {
		t.Fatal(err)
	}
	names := []string{
		"20260713010000_create_reject_mutation_routine",
		"20260713014000_create_tenant_migration_checkpoints",
		"20260713014100_create_checkpoint_transition_routine",
		"20260713014200_guard_checkpoint_transitions",
		"20260714010500_create_tenant_checkpoint_attempts",
		"20260714010800_refine_tenant_checkpoint_transition",
		"20260714010900_guard_tenant_checkpoint_transitions_v2",
		"20260730120300_create_tenant_migration_supersession_registry_v1",
		"20260730120301_allow_failed_tenant_checkpoint_supersession_v1",
		"20260730120302_create_tenant_checkpoint_requirement_matcher_v1",
		"20260730120303_create_tenant_requirement_verifier_v3",
		"20260730120304_assert_tenant_supersession_release_frontier_v1",
	}
	names = append(names, tenantSupersessionChainMigrations...)
	for _, name := range names {
		source := readIAMNamedPlatformMigration(t, "fresh", name)
		if _, err := database.ExecContext(ctx, string(source)); err != nil {
			t.Fatalf("apply %s: %v", name, err)
		}
	}
}

func assertCrossOwnerSupersessionEdgeRejected(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `INSERT INTO
platform.tenant_migration_supersessions (
 owner,scope,logical_version,database_origin,superseded_path,
 superseded_sha256,current_path,current_sha256,supersession_distance
) VALUES ('other_owner','tenant',$1,'fresh',$2,$3,$4,$5,3)`,
		chainVersion, chainOriginal, chainOriginalSH, chainTerminal, chainTerminalSH)
	requireChainSQLState(t, err, "23505")
}

func seedFailedChainCheckpoint(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenantID, selectedPath, selectedSHA string,
) {
	t.Helper()
	setChainCheckpointTrigger(t, ctx, database, false)
	_, err := database.ExecContext(ctx, `INSERT INTO
platform.tenant_migration_checkpoints (
 tenant_id,owner,logical_version,database_origin,selected_path,
 selected_sha256,run_id,state,attempt,started_at,finished_at,error_code
) VALUES ($1,'system_resource_layout',$2,'fresh',$3,$4,$5,'failed',1,
 pg_catalog.clock_timestamp()-INTERVAL '2 seconds',
 pg_catalog.clock_timestamp()-INTERVAL '1 second','migration_execution_failed')`,
		tenantID, chainVersion, selectedPath, selectedSHA, chainOldRunID)
	if err != nil {
		t.Fatal(err)
	}
	setChainCheckpointTrigger(t, ctx, database, true)
}

func transitionChainCheckpoint(
	t *testing.T, ctx context.Context, database *sql.DB, tenantID string,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `UPDATE platform.tenant_migration_checkpoints
SET selected_path=$2,selected_sha256=$3,run_id=$4,state='running',attempt=2
WHERE tenant_id=$1 AND owner='system_resource_layout' AND logical_version=$5`,
		tenantID, chainTerminal, chainTerminalSH, chainNewRunID, chainVersion)
	if err != nil {
		t.Fatalf("transition chain checkpoint %s: %v", tenantID, err)
	}
	var state, path string
	err = database.QueryRowContext(ctx, `SELECT state,selected_path
FROM platform.tenant_migration_checkpoints
WHERE tenant_id=$1 AND owner='system_resource_layout'`, tenantID).Scan(&state, &path)
	if err != nil || state != "running" || path != chainTerminal {
		t.Fatalf("chain checkpoint state=%s path=%s: %v", state, path, err)
	}
}

func setChainCheckpointTrigger(
	t *testing.T, ctx context.Context, database *sql.DB, enabled bool,
) {
	t.Helper()
	statement := `ALTER TABLE platform.tenant_migration_checkpoints
DISABLE TRIGGER tenant_migration_checkpoints_validate_transition_v5`
	if enabled {
		statement = `ALTER TABLE platform.tenant_migration_checkpoints
ENABLE ALWAYS TRIGGER tenant_migration_checkpoints_validate_transition_v5`
	}
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatal(err)
	}
}

func requireChainSQLState(t *testing.T, err error, expected string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != expected {
		t.Fatalf("SQL error = %v, want %s", err, expected)
	}
}
