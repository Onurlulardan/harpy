//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertDocumentsFormsCleanupEnqueueReplay(t *testing.T, ctx context.Context,
	tx *sql.Tx, operationID, itemID, jobID, competingJobID string, now time.Time,
) {
	t.Helper()
	runAt := now.Add(time.Microsecond)
	execDocumentsSQL(t, ctx, tx, documentsFormsCleanupSQLRoot+"stage_enqueue.sql",
		operationID, itemID, jobID, runAt)
	assertDocumentsFormsCleanupStagedJob(t, ctx, tx, operationID, itemID, jobID, runAt)

	execDocumentsSQL(t, ctx, tx, documentsFormsCleanupSQLRoot+"prepare.sql",
		operationID, now)
	execDocumentsSQL(t, ctx, tx, documentsFormsCleanupSQLRoot+"stage_enqueue.sql",
		operationID, itemID, competingJobID, runAt.Add(time.Minute))
	assertDocumentsFormsCleanupStagedJob(t, ctx, tx, operationID, itemID, jobID, runAt)

	execDocumentsSQL(t, ctx, tx, documentsFormsCleanupSQLRoot+"mark_enqueued.sql",
		operationID, itemID, jobID, runAt, now)
}

func assertDocumentsFormsCleanupStagedJob(t *testing.T, ctx context.Context,
	tx *sql.Tx, operationID, itemID, jobID string, runAt time.Time,
) {
	t.Helper()
	var storedJob string
	var storedRunAt time.Time
	var enqueuedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT purge_job_id,purge_run_at,enqueued_at
 FROM documents_forms_cleanup_targets WHERE operation_id=$1 AND item_id=$2`,
		operationID, itemID).Scan(&storedJob, &storedRunAt, &enqueuedAt)
	if err != nil || storedJob != jobID || !storedRunAt.Equal(runAt) ||
		enqueuedAt.Valid {
		t.Fatalf("cleanup staged job = %s/%s/%v/%v", storedJob, storedRunAt, enqueuedAt, err)
	}
}
