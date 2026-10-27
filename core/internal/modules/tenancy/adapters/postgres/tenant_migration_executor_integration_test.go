//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestTenantMigrationExecutorLifecycleAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startTenantMigrationDatabase(t, ctx)
	for _, origin := range []string{"fresh"} {
		t.Run(origin+"/complete", func(t *testing.T) {
			_, fixture, claim := prepareTenantExecutorScenario(
				t, ctx, database, origin,
			)
			heartbeatTenantMigration(t, ctx, database, claim)
			createTenantSchemaReceipt(t, ctx, database, claim)
			assertSchemaResumeEvidence(t, ctx, database, claim)
			seedCompletedTenantRun(t, ctx, database, claim, origin)
			completeTenantMigration(t, ctx, database, claim)
			assertTenantMigrationReady(t, ctx, database, fixture.operationID)
			assertTenantRemainsInvisible(t, ctx, database, fixture.tenantID)
		})
		t.Run(origin+"/lease-expiry", func(t *testing.T) {
			profileID, _, claim := prepareTenantExecutorScenario(
				t, ctx, database, origin,
			)
			assertExpiredLeaseReclaims(t, ctx, database, profileID, claim)
		})
		t.Run(origin+"/offline-cleanup", func(t *testing.T) {
			profileID, _, claim := prepareTenantExecutorScenario(
				t, ctx, database, origin,
			)
			assertOrphanRequiresOfflineCleanup(
				t, ctx, database, profileID, claim,
			)
		})
		t.Run(origin+"/generation-fence", func(t *testing.T) {
			profileID, _, jobID := prepareTenantExecutorPendingScenario(
				t, ctx, database, origin,
			)
			assertStaleGenerationFences(t, ctx, database, profileID, jobID)
		})
	}
}

func assertTenantMigrationReady(
	t *testing.T, ctx context.Context, database *sql.DB, operationID uuid.UUID,
) {
	t.Helper()
	var state string
	if err := database.QueryRowContext(ctx, `SELECT state FROM
platform.tenancy_onboarding_operations WHERE operation_id=$1`, operationID).
		Scan(&state); err != nil || state != "migration_ready" {
		t.Fatalf("tenant migration onboarding state = %q, err=%v", state, err)
	}
}

func TestTenantMigrationExecutorRejectsPersistedModeMismatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startTenantMigrationDatabase(t, ctx)
	resetTenantMigrationCatalog(t, ctx, database)
	runPlatformMigrations(t, ctx, database, "fresh")
	insertDeploymentPolicy(t, ctx, database, "saas")
	prepareTenantExecutorFoundation(t, ctx, database, "fresh")
	var profileID uuid.UUID
	err := database.QueryRowContext(ctx, `SELECT profile_instance_id FROM
platform.tenancy_tenant_migration_ready_register_v1($1,$2,$3,$4,$5,$6)`,
		uuid.New(), tenantExecutorReleaseID, "onpremise", 11,
		strings.Repeat("a", 64), 30).Scan(&profileID)
	if err == nil || !strings.Contains(err.Error(), "policy conflicts") {
		t.Fatalf("persisted mode mismatch error = %v", err)
	}
}

func TestTenantMigrationExecutorCompletesSaaSEnvironmentAuthorityTransition(
	t *testing.T,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startTenantMigrationDatabase(t, ctx)
	resetTenantMigrationCatalog(t, ctx, database)
	runPlatformMigrations(t, ctx, database, "fresh")
	prepareTenantExecutorFoundation(t, ctx, database, "fresh")
	profileID := uuid.New()
	registerTenantExecutorMode(t, ctx, database, profileID, "saas")
	insertDeploymentPolicy(t, ctx, database, "saas")
	fixture := acceptTenantExecutorOperationMode(t, ctx, database, "saas")
	jobID := enqueueTenantMigration(t, ctx, database, fixture.operationID)
	claim := claimTenantMigration(t, ctx, database, profileID, jobID)
	createTenantSchemaReceipt(t, ctx, database, claim)
	seedCompletedTenantRun(t, ctx, database, claim, "fresh")
	completeTenantMigration(t, ctx, database, claim)
	var state string
	if err := database.QueryRowContext(ctx, `SELECT state FROM
platform.tenancy_onboarding_operations WHERE operation_id=$1`,
		fixture.operationID).Scan(&state); err != nil || state != "authority_pending" {
		t.Fatalf("SaaS environment migration state = %q, err=%v", state, err)
	}
}

func prepareTenantExecutorScenario(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) (uuid.UUID, onboardingFixture, executorClaim) {
	t.Helper()
	profileID, fixture, jobID := prepareTenantExecutorPendingScenario(
		t, ctx, database, origin,
	)
	claim := claimTenantMigration(t, ctx, database, profileID, jobID)
	return profileID, fixture, claim
}

func prepareTenantExecutorPendingScenario(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) (uuid.UUID, onboardingFixture, uuid.UUID) {
	t.Helper()
	resetTenantMigrationCatalog(t, ctx, database)
	runPlatformMigrations(t, ctx, database, origin)
	prepareTenantExecutorFoundation(t, ctx, database, origin)
	profileID := uuid.New()
	registerTenantExecutor(t, ctx, database, profileID)
	insertDeploymentPolicy(t, ctx, database, "onpremise")
	fixture := acceptTenantExecutorOperation(t, ctx, database)
	jobID := enqueueTenantMigration(t, ctx, database, fixture.operationID)
	return profileID, fixture, jobID
}

func registerTenantExecutor(
	t *testing.T, ctx context.Context, database *sql.DB, profileID uuid.UUID,
) {
	t.Helper()
	registerTenantExecutorMode(t, ctx, database, profileID, "onpremise")
}

func registerTenantExecutorMode(
	t *testing.T, ctx context.Context, database *sql.DB,
	profileID uuid.UUID, mode string,
) {
	t.Helper()
	var returned, releaseID, environmentID uuid.UUID
	var returnedMode, digest string
	var runtimeGeneration, databaseGeneration int64
	var readyUntil time.Time
	err := database.QueryRowContext(ctx, `SELECT * FROM
platform.tenancy_tenant_migration_ready_register_v1($1,$2,$3,$4,$5,$6)`,
		profileID, tenantExecutorReleaseID, mode, 11,
		strings.Repeat("a", 64), 30).Scan(&returned, &releaseID, &environmentID,
		&returnedMode, &runtimeGeneration, &databaseGeneration, &digest, &readyUntil)
	if err != nil || returned != profileID || !readyUntil.After(time.Now()) {
		t.Fatalf("register tenant executor: ready=%s err=%v", readyUntil, err)
	}
}

func enqueueTenantMigration(
	t *testing.T, ctx context.Context, database *sql.DB,
	operationID uuid.UUID,
) uuid.UUID {
	t.Helper()
	jobID := uuid.New()
	var returned, operation, tenant uuid.UUID
	var state, schema, step string
	err := database.QueryRowContext(ctx, `SELECT * FROM
platform.tenancy_tenant_migration_enqueue_v1($1,$2,$3,$4,$5,$6)`,
		jobID, operationID, tenantExecutorReleaseID, 11, 7,
		strings.Repeat("a", 64)).Scan(&returned, &operation, &state, &tenant,
		&schema, &step)
	if err != nil || returned != jobID || state != "pending" {
		t.Fatalf("enqueue tenant migration: %s/%s %v", returned, state, err)
	}
	return jobID
}

func quoteSchema(schema string) string { return pgx.Identifier{schema}.Sanitize() }
