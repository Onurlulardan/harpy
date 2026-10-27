//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertActiveTerminalFailureAccepted(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	jobID string,
	code string,
) {
	t.Helper()
	finished := time.Now().UTC().Add(-time.Millisecond).Truncate(time.Microsecond)
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	detail := integrationDigest(jobID + "-active-detail")
	if _, err := tx.ExecContext(ctx, schemaAttemptInsertSQL,
		jobID, 0, finished, "failed", integrationDigest(jobID+"-active-outcome"),
		code, detail); err != nil {
		t.Fatalf("active terminal failure receipt rejected: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE platform.schema_job_steps SET
generation=generation+1,state='failed',next_executor='none',finished_at=$2,
error_code=$3,updated_at=$2 WHERE job_id=$1 AND ordinal=0`,
		jobID, finished, code); err != nil {
		t.Fatalf("active terminal step failure rejected: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE platform.schema_jobs SET
generation=generation+1,state='failed',next_executor='none',
lease_id=NULL,lease_owner_id=NULL,leased_at=NULL,lease_heartbeat_at=NULL,
lease_expires_at=NULL,finished_at=$2,failure_code=$3,updated_at=$2
WHERE job_id=$1`, jobID, finished, code); err != nil {
		t.Fatalf("active terminal job failure rejected: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("active terminal failure commit rejected: %v", err)
	}
}
