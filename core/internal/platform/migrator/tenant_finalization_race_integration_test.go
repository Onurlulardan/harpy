//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

const raceTenantID = "71717171-0000-4000-8000-000000000071"

func TestHistoricalEmptyTenantSelectionFinalizationPostgreSQL184FailsClosed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database, runner, connection := stageTenantFinalizationRace(t, ctx)
	runner.tenantFiles, runner.tenantSelected = nil, nil
	target := postgresstore.TenantMigrationTarget{
		TenantID:   "72727272-0000-4000-8000-000000000072",
		SchemaName: "tenant_historical_empty_selection",
	}
	insertTenantTarget(t, ctx, database, target)
	err := runner.finalizeSuccess(ctx, connection, runnerIntegrationRunID)
	if err == nil || !strings.Contains(err.Error(), "tenant migration selection is absent") {
		t.Fatalf("historical empty selection finalization error = %v", err)
	}
	assertTenantRaceFinalizationAbsent(t, ctx, database)
}

func TestTenantFinalizationPostgreSQL184WaitsForActiveInsertAndFailsClosed(t *testing.T) {
	runTenantFinalizationRace(t, func(
		t *testing.T, ctx context.Context, database *sql.DB,
	) *sql.Tx {
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		insertRaceTenant(t, ctx, transaction, raceTenantID, "tenant_race_insert")
		return transaction
	})
}

func TestTenantFinalizationPostgreSQL184WaitsForReactivationAndFailsClosed(t *testing.T) {
	runTenantFinalizationRace(t, func(
		t *testing.T, ctx context.Context, database *sql.DB,
	) *sql.Tx {
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		insertRaceTenant(t, ctx, transaction, raceTenantID, "tenant_race_reactivate")
		if _, err := transaction.ExecContext(
			ctx, "UPDATE platform.tenants SET deleted_at=pg_catalog.clock_timestamp() WHERE id=$1",
			raceTenantID,
		); err != nil {
			t.Fatal(err)
		}
		if err := transaction.Commit(); err != nil {
			t.Fatal(err)
		}
		transaction, err = database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := transaction.ExecContext(
			ctx, "UPDATE platform.tenants SET deleted_at=NULL WHERE id=$1", raceTenantID,
		); err != nil {
			t.Fatal(err)
		}
		return transaction
	})
}

func runTenantFinalizationRace(
	t *testing.T,
	startWriter func(*testing.T, context.Context, *sql.DB) *sql.Tx,
) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database, runner, connection := stageTenantFinalizationRace(t, ctx)
	applied, err := runner.runTenantMigrations(ctx, connection, runnerIntegrationRunID)
	if err != nil || len(applied) != 0 {
		t.Fatalf("initial empty tenant enumeration = %#v err=%v", applied, err)
	}
	writer := startWriter(t, ctx, database)
	defer writer.Rollback()
	finalized := make(chan error, 1)
	go func() {
		finalized <- runner.finalizeSuccess(ctx, connection, runnerIntegrationRunID)
	}()
	waitForTenantShareLock(t, ctx, database, finalized)
	if err := writer.Commit(); err != nil {
		t.Fatalf("commit concurrent tenant mutation: %v", err)
	}
	err = <-finalized
	if err == nil || !strings.Contains(err.Error(), "tenant checkpoint") {
		t.Fatalf("finalization race error = %v", err)
	}
	assertTenantRaceFinalizationAbsent(t, ctx, database)
}

func waitForTenantShareLock(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	finalized <-chan error,
) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var waiting bool
		err := database.QueryRowContext(ctx, `SELECT EXISTS (
SELECT 1 FROM pg_catalog.pg_locks
WHERE relation='platform.tenants'::regclass AND mode='ShareLock' AND NOT granted
)`).Scan(&waiting)
		if err != nil {
			t.Fatal(err)
		}
		if waiting {
			return
		}
		select {
		case err := <-finalized:
			t.Fatalf("finalization did not wait for tenant writer: %v", err)
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("finalization tenant SHARE lock was not observed")
}

func assertTenantRaceFinalizationAbsent(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var state string
	if err := database.QueryRowContext(ctx,
		"SELECT state FROM platform.migration_runs WHERE run_id=$1",
		runnerIntegrationRunID,
	).Scan(&state); err != nil || state != "running" {
		t.Fatalf("race run state = %q err=%v", state, err)
	}
	var releases, ready int
	if err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM platform.release_generations WHERE migration_run_id=$1),
(SELECT count(*) FROM platform.schema_contracts WHERE component='core' AND status='ready')`,
		runnerIntegrationRunID,
	).Scan(&releases, &ready); err != nil || releases != 0 || ready != 0 {
		t.Fatalf("race release/ready = %d/%d err=%v", releases, ready, err)
	}
}
