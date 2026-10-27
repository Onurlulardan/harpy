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

func TestTenantMigrationExecutorClaimIdlesBeforePolicyAgainstPostgreSQL184(
	t *testing.T,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startTenantMigrationDatabase(t, ctx)
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			resetTenantMigrationCatalog(t, ctx, database)
			if origin == "v1_adopted" {
				prepareV1TenantRegistry(t, ctx, database)
			}
			runPlatformMigrations(t, ctx, database, origin)
			prepareTenantExecutorFoundation(t, ctx, database, origin)
			profileID := uuid.New()
			registerTenantExecutor(t, ctx, database, profileID)
			assertTenantExecutorHeartbeatBeforePolicy(t, ctx, database, profileID)
			assertTenantExecutorClaimIsIdle(t, ctx, database, profileID)
			insertDeploymentPolicy(t, ctx, database, "saas")
			assertTenantExecutorHeartbeatRejectsPolicyConflict(
				t, ctx, database, profileID,
			)
		})
	}
}

func assertTenantExecutorHeartbeatBeforePolicy(
	t *testing.T, ctx context.Context, database *sql.DB, profileID uuid.UUID,
) {
	t.Helper()
	var readyUntil time.Time
	err := database.QueryRowContext(ctx, `SELECT
platform.tenancy_tenant_migration_ready_heartbeat_v1($1,$2)`,
		profileID, 30).Scan(&readyUntil)
	if err != nil || !readyUntil.After(time.Now()) {
		t.Fatalf("pre-policy heartbeat ready=%s error=%v", readyUntil, err)
	}
}

func assertTenantExecutorHeartbeatRejectsPolicyConflict(
	t *testing.T, ctx context.Context, database *sql.DB, profileID uuid.UUID,
) {
	t.Helper()
	var readyUntil time.Time
	err := database.QueryRowContext(ctx, `SELECT
platform.tenancy_tenant_migration_ready_heartbeat_v1($1,$2)`,
		profileID, 30).Scan(&readyUntil)
	if err == nil || !strings.Contains(err.Error(), "readiness is stale") {
		t.Fatalf("conflicting-policy heartbeat ready=%s error=%v", readyUntil, err)
	}
}

func assertTenantExecutorClaimIsIdle(
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
		t.Fatalf("pre-policy claim error = %v, want no rows", err)
	}
}
