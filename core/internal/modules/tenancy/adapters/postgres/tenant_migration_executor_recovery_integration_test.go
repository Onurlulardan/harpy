//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func assertExpiredLeaseReclaims(
	t *testing.T, ctx context.Context, database *sql.DB, profileID uuid.UUID,
	first executorClaim,
) {
	t.Helper()
	expireTenantMigrationLease(t, ctx, database, first)
	second := claimTenantMigration(t, ctx, database, profileID, first.jobID)
	if second.attempt != 2 || second.previousRunID != first.runID {
		t.Fatalf("expired lease did not fence prior run: %+v", second)
	}
}

func assertStaleGenerationFences(
	t *testing.T, ctx context.Context, database *sql.DB,
	profileID, jobID uuid.UUID,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `UPDATE platform.release_generations
SET runtime_generation=12 WHERE release_id=$1`, tenantExecutorReleaseID); err != nil {
		t.Fatal(err)
	}
	connection := tenantMigratorConnection(t, ctx, database)
	defer connection.Close()
	var count int
	err := connection.QueryRowContext(ctx, `SELECT count(*) FROM
platform.tenancy_tenant_migration_claim_v1($1,$2,$3,$4,$5)`,
		profileID, uuid.New(), uuid.New(), uuid.New(), 30).Scan(&count)
	if err == nil {
		t.Fatalf("stale release generation claimed job %s", jobID)
	}
	var state string
	if readErr := database.QueryRowContext(ctx, `SELECT state FROM
platform.tenancy_tenant_migration_jobs WHERE job_id=$1`, jobID).Scan(&state); readErr != nil || state != "pending" {
		t.Fatalf("generation fence mutated job: state=%s err=%v", state, readErr)
	}
}

func assertOrphanRequiresOfflineCleanup(
	t *testing.T, ctx context.Context, database *sql.DB, profileID uuid.UUID,
	claim executorClaim,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx,
		"CREATE SCHEMA "+quoteSchema(claim.schema)+" AUTHORIZATION harpy_owner"); err != nil {
		t.Fatal(err)
	}
	var tenantID uuid.UUID
	var schema string
	var namespaceExists, receiptExists bool
	err := database.QueryRowContext(ctx, `SELECT * FROM
platform.tenancy_tenant_migration_schema_inspect_v1($1,$2,$3,$4)`,
		claim.jobID, claim.attempt, claim.profileID, claim.leaseID).
		Scan(&tenantID, &schema, &namespaceExists, &receiptExists)
	if err != nil || !namespaceExists || receiptExists {
		t.Fatalf("orphan schema evidence=%t/%t err=%v",
			namespaceExists, receiptExists, err)
	}
	digest := jobDigest(claim.jobID, "cleanup-required")
	var recorded bool
	err = database.QueryRowContext(ctx, `SELECT
platform.tenancy_tenant_migration_fail_v1($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		claim.jobID, claim.attempt, profileID, claim.leaseID, "cleanup_required",
		"orphan_schema_requires_cleanup", uuid.New(), digest, digest).Scan(&recorded)
	if err != nil || !recorded {
		t.Fatalf("record cleanup requirement: %v", err)
	}
	retireTenantExecutor(t, ctx, database, profileID)
	if _, err := database.ExecContext(ctx, "DROP SCHEMA "+quoteSchema(schema)+" CASCADE"); err != nil {
		t.Fatal(err)
	}
	offlineDigest := jobDigest(claim.jobID, "offline-cleanup")
	err = database.QueryRowContext(ctx, `SELECT
platform.tenancy_tenant_migration_offline_cleanup_record_v1($1,$2,$3,$4)`,
		claim.jobID, uuid.New(), offlineDigest, offlineDigest).Scan(&recorded)
	if err != nil || !recorded {
		t.Fatalf("record offline cleanup: %v", err)
	}
	var state string
	if err := database.QueryRowContext(ctx,
		"SELECT state FROM platform.tenancy_tenant_migration_jobs WHERE job_id=$1",
		claim.jobID).Scan(&state); err != nil || state != "fenced" {
		t.Fatalf("offline cleanup state=%s err=%v", state, err)
	}
}

func retireTenantExecutor(
	t *testing.T, ctx context.Context, database *sql.DB, profileID uuid.UUID,
) {
	t.Helper()
	var retired bool
	err := database.QueryRowContext(ctx, `SELECT
platform.tenancy_tenant_migration_ready_retire_v1($1)`, profileID).Scan(&retired)
	if err != nil || !retired {
		t.Fatalf("retire tenant executor: %v", err)
	}
}

func jobDigest(jobID uuid.UUID, suffix string) string {
	sum := sha256.Sum256([]byte(jobID.String() + ":" + suffix))
	return strings.ToLower(hex.EncodeToString(sum[:]))
}
