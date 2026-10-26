//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	rollforwardReleaseID = "20000000-0000-4000-8000-000000000100"
	rollforwardDeployRun = "30000000-0000-4000-8000-000000000100"
)

func TestTenantMigrationExecutorRollsAttemptLineageToNewReleasePG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startTenantMigrationDatabase(t, ctx)
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			profileID, _, jobID := prepareTenantExecutorPendingScenario(
				t, ctx, database, origin,
			)
			claim := claimTenantMigration(t, ctx, database, profileID, jobID)
			for claim.attempt < 8 {
				expireTenantMigrationLease(t, ctx, database, claim)
				claim = claimTenantMigration(t, ctx, database, profileID, jobID)
			}
			expireTenantMigrationLease(t, ctx, database, claim)
			assertReleaseAttemptBudgetExhausted(
				t, ctx, database, profileID,
			)

			nextProfileID := registerRollforwardRelease(
				t, ctx, database, origin,
			)
			rolled := claimTenantMigration(
				t, ctx, database, nextProfileID, jobID,
			)
			if rolled.attempt != 9 || rolled.previousRunID != claim.runID {
				t.Fatalf("roll-forward claim = %+v", rolled)
			}
			assertRollforwardJobIdentity(t, ctx, database, jobID)
		})
	}
}

func assertReleaseAttemptBudgetExhausted(
	t *testing.T, ctx context.Context, database *sql.DB, profileID uuid.UUID,
) {
	t.Helper()
	connection := tenantMigratorConnection(t, ctx, database)
	defer connection.Close()
	var jobID uuid.UUID
	err := connection.QueryRowContext(ctx, `SELECT job_id FROM
platform.tenancy_tenant_migration_claim_v1($1,$2,$3,$4,$5)`,
		profileID, uuid.New(), uuid.New(), uuid.New(), 30).Scan(&jobID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("ninth same-release claim error = %v, want no rows", err)
	}
}

func registerRollforwardRelease(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) uuid.UUID {
	t.Helper()
	digest := strings.Repeat("f", 64)
	execTenantExecutorFixture(t, ctx, database, `INSERT INTO platform.migration_runs (
run_id,release_id,environment_id,database_origin,command,state,
migration_manifest_sha256,extension_manifest_sha256,roles_manifest_sha256,
privileges_manifest_sha256,ownership_manifest_sha256,started_at,finished_at)
VALUES ($1,$2,$3,$4,'up','succeeded',$5,$5,$5,$5,$5,now(),now())`,
		rollforwardDeployRun, rollforwardReleaseID, testEnvironmentID, origin, digest)
	execTenantExecutorFixture(t, ctx, database,
		`INSERT INTO platform.release_generations VALUES ($1,$2,$3,$4,8,12,$5)`,
		rollforwardDeployRun, rollforwardReleaseID, testEnvironmentID, origin, digest)
	execTenantExecutorFixture(t, ctx, database, `UPDATE platform.schema_contracts
SET verified_run_id=$1,database_generation=8,migration_manifest_sha256=$2
WHERE component='core'`, rollforwardDeployRun, digest)
	profileID := uuid.New()
	var returned uuid.UUID
	err := database.QueryRowContext(ctx, `SELECT profile_instance_id FROM
platform.tenancy_tenant_migration_ready_register_v1($1,$2,$3,$4,$5,$6)`,
		profileID, rollforwardReleaseID, "onpremise", 12, digest, 120).
		Scan(&returned)
	if err != nil || returned != profileID {
		t.Fatalf("register roll-forward release: %s %v", returned, err)
	}
	return profileID
}

func assertRollforwardJobIdentity(
	t *testing.T, ctx context.Context, database *sql.DB, jobID uuid.UUID,
) {
	t.Helper()
	var releaseID uuid.UUID
	var runtimeGeneration, databaseGeneration int64
	var digest string
	var attempt, attempts int
	err := database.QueryRowContext(ctx, `SELECT job.release_id,
job.runtime_generation,job.database_generation,job.migration_manifest_sha256,
job.current_attempt,(SELECT count(*) FROM
platform.tenancy_tenant_migration_attempts AS attempt
WHERE attempt.job_id=job.job_id)
FROM platform.tenancy_tenant_migration_jobs AS job WHERE job.job_id=$1`, jobID).
		Scan(&releaseID, &runtimeGeneration, &databaseGeneration,
			&digest, &attempt, &attempts)
	if err != nil || releaseID.String() != rollforwardReleaseID ||
		runtimeGeneration != 12 || databaseGeneration != 8 ||
		digest != strings.Repeat("f", 64) || attempt != 9 || attempts != 9 {
		t.Fatalf("roll-forward identity = %s/%d/%d/%s/%d/%d err=%v",
			releaseID, runtimeGeneration, databaseGeneration,
			digest, attempt, attempts, err)
	}
}
