//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertExpiredExhaustionAccepted(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	jobID string,
) {
	t.Helper()
	finished := time.Now().UTC().Add(-time.Millisecond).Truncate(time.Microsecond)
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	code := "attempt_limit_exhausted"
	detail := integrationDigest(jobID + "-detail")
	if _, err := tx.ExecContext(ctx, schemaAttemptInsertSQL,
		jobID, 0, finished, "failed", integrationDigest(jobID+"-outcome"),
		code, detail); err != nil {
		t.Fatalf("expired attempt exhaustion receipt rejected: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE platform.schema_job_steps SET
generation=generation+1,state='failed',next_executor='none',finished_at=$2,
error_code=$3,updated_at=$2 WHERE job_id=$1 AND ordinal=0`,
		jobID, finished, code); err != nil {
		t.Fatalf("expired step exhaustion rejected: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE platform.schema_jobs SET
generation=generation+1,state='failed',next_executor='none',
lease_id=NULL,lease_owner_id=NULL,leased_at=NULL,lease_heartbeat_at=NULL,
lease_expires_at=NULL,finished_at=$2,failure_code=$3,updated_at=$2
WHERE job_id=$1`, jobID, finished, code); err != nil {
		t.Fatalf("expired job exhaustion rejected: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("expired exhaustion commit rejected: %v", err)
	}
}

func assertStaleSchemaTransitionsRejected(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	assertSchemaUpdateRejected(t, ctx, database, `UPDATE platform.schema_job_steps
SET updated_at=pg_catalog.clock_timestamp() WHERE job_id=$1 AND ordinal=0`, progressJobID)
	assertSchemaUpdateRejected(t, ctx, database, `UPDATE platform.schema_job_steps SET
generation=generation+1,checkpoint_record_id=$2,processed_rows=processed_rows+1001,
updated_at=pg_catalog.clock_timestamp() WHERE job_id=$1 AND ordinal=0`,
		progressJobID, "018f0000-0000-7000-8000-000000000712")
	assertSchemaUpdateRejected(t, ctx, database, `UPDATE platform.schema_jobs
SET updated_at=pg_catalog.clock_timestamp() WHERE job_id=$1`, progressJobID)
}

func assertSchemaUpdateRejected(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	query string,
	arguments ...any,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, query, arguments...); err == nil {
		t.Fatalf("stale or illegal schema transition was accepted")
	}
}
