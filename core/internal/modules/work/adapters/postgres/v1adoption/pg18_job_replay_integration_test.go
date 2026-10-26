//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func assertJobReceiptCrashReplay(t *testing.T, ctx context.Context,
	database *sql.DB, schema string,
) {
	t.Helper()
	jobID := uuid.MustParse("019b2ee0-1800-7a1a-8d21-5ca5797f2800")
	crashed, err := beginWorkSchema(ctx, database, schema)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = crashed.ExecContext(ctx, workBeginReceiptSQL, jobID); err != nil {
		_ = crashed.Rollback()
		t.Fatal(err)
	}
	if err = crashed.Rollback(); err != nil {
		t.Fatal(err)
	}
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		var count int
		if scanErr := transaction.QueryRowContext(ctx,
			"SELECT count(*) FROM work_job_receipts WHERE job_id=$1", jobID).Scan(&count); scanErr != nil {
			t.Fatal(scanErr)
		}
		if count != 0 {
			t.Fatalf("crashed receipt count = %d", count)
		}
	})
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		if _, executeErr := transaction.ExecContext(ctx, workBeginReceiptSQL, jobID); executeErr != nil {
			t.Fatal(executeErr)
		}
		if _, executeErr := transaction.ExecContext(ctx, `
UPDATE work_job_receipts SET state='SUCCEEDED', outcome_sha256=decode(repeat('02',32),'hex'),
 completed_at=transaction_timestamp()
WHERE job_id=$1 AND state='RUNNING'`, jobID); executeErr != nil {
			t.Fatal(executeErr)
		}
	})
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		result, executeErr := transaction.ExecContext(ctx, workBeginReceiptSQL, jobID)
		if executeErr != nil {
			t.Fatal(executeErr)
		}
		affected, _ := result.RowsAffected()
		var state string
		var generation int64
		if scanErr := transaction.QueryRowContext(ctx, `
SELECT state,lease_generation FROM work_job_receipts WHERE job_id=$1`, jobID).
			Scan(&state, &generation); scanErr != nil {
			t.Fatal(scanErr)
		}
		if affected != 0 || state != "SUCCEEDED" || generation != 2 {
			t.Fatalf("replay affected=%d state=%s generation=%d", affected, state, generation)
		}
	})
}

const workBeginReceiptSQL = `
INSERT INTO work_job_receipts
(job_id,job_kind,payload_sha256,lease_generation,state,outcome_sha256,started_at,completed_at)
VALUES ($1,'work.reminder.dispatch',decode(repeat('01',32),'hex'),2,'RUNNING',NULL,
 transaction_timestamp(),NULL)
ON CONFLICT (job_id) DO NOTHING`
