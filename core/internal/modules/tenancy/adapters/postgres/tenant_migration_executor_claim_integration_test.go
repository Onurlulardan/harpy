//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func claimTenantMigration(
	t *testing.T, ctx context.Context, database *sql.DB,
	profileID, expectedJobID uuid.UUID,
) executorClaim {
	t.Helper()
	connection := tenantMigratorConnection(t, ctx, database)
	defer connection.Close()
	leaseID, runID, expiryID := uuid.New(), uuid.New(), uuid.New()
	claim := executorClaim{leaseID: leaseID, runID: runID, profileID: profileID}
	var step, origin, mode, digest string
	var releaseID, environmentID uuid.UUID
	var previous *uuid.UUID
	var runtimeGeneration, databaseGeneration int64
	var leaseUntil time.Time
	err := connection.QueryRowContext(ctx, `SELECT * FROM
platform.tenancy_tenant_migration_claim_v1($1,$2,$3,$4,$5)`,
		profileID, leaseID, runID, expiryID, 30).Scan(&claim.jobID,
		&claim.operationID, &claim.tenantID, &claim.schema, &step, &origin,
		&claim.attempt, &claim.runID, &previous, &claim.leaseID, &claim.profileID,
		&releaseID, &environmentID, &mode, &runtimeGeneration,
		&databaseGeneration, &digest, &leaseUntil)
	if previous != nil {
		claim.previousRunID = *previous
	}
	if err != nil || claim.jobID != expectedJobID || origin != "new" ||
		claim.runID != runID || !leaseUntil.After(time.Now()) {
		t.Fatalf("claim tenant migration: %+v err=%v", claim, err)
	}
	return claim
}

func tenantMigratorConnection(
	t *testing.T, ctx context.Context, database *sql.DB,
) *sql.Conn {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := connection.ExecContext(
		ctx, `SET SESSION AUTHORIZATION harpy_migrator`,
	); err != nil {
		connection.Close()
		t.Fatalf("set migrator session authorization: %v", err)
	}
	if _, err := connection.ExecContext(ctx, `SET ROLE harpy_owner`); err != nil {
		connection.Close()
		t.Fatalf("set owner role: %v", err)
	}
	return connection
}

func heartbeatTenantMigration(
	t *testing.T, ctx context.Context, database *sql.DB, claim executorClaim,
) {
	t.Helper()
	var leaseUntil time.Time
	err := database.QueryRowContext(ctx, `SELECT
platform.tenancy_tenant_migration_heartbeat_v1($1,$2,$3,$4,$5)`,
		claim.jobID, claim.attempt, claim.profileID, claim.leaseID, 30).
		Scan(&leaseUntil)
	if err != nil || !leaseUntil.After(time.Now()) {
		t.Fatalf("heartbeat tenant migration: %s %v", leaseUntil, err)
	}
}

func expireTenantMigrationLease(
	t *testing.T, ctx context.Context, database *sql.DB, claim executorClaim,
) {
	t.Helper()
	statements := []string{
		"ALTER TABLE platform.tenancy_tenant_migration_attempts DISABLE TRIGGER USER",
		"UPDATE platform.tenancy_tenant_migration_attempts SET heartbeat_at = leased_at, lease_until = leased_at + interval '1 microsecond' WHERE job_id = $1 AND attempt = $2",
		"ALTER TABLE platform.tenancy_tenant_migration_attempts ENABLE TRIGGER USER",
	}
	for index, statement := range statements {
		arguments := []any(nil)
		if index == 1 {
			arguments = []any{claim.jobID, claim.attempt}
		}
		if _, err := database.ExecContext(ctx, statement, arguments...); err != nil {
			t.Fatalf("expire tenant lease: %v", err)
		}
	}
}
