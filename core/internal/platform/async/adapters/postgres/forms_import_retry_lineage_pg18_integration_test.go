//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func assertFormsImportRetryLineage(
	t *testing.T, ctx context.Context, database *sql.DB, runtime *pgxpool.Pool,
	call formsImportRetryCall, result formsImportRetryResult,
) {
	t.Helper()
	var state string
	var cancelled, changed bool
	var attempts, maxAttempts int
	var completedAt any
	err := runtime.QueryRow(ctx, `SELECT * FROM public.read_forms_import_job(
 $1,$2,$3)`, call.tenantID, call.importID, result.jobID).Scan(
		&state, &cancelled, &attempts, &maxAttempts, &completedAt)
	if err != nil || state != "queued" || cancelled || attempts != 0 ||
		maxAttempts != 8 || completedAt != nil {
		t.Fatalf("retry read = %s %t %d/%d %v, %v",
			state, cancelled, attempts, maxAttempts, completedAt, err)
	}
	err = runtime.QueryRow(ctx, `SELECT * FROM public.control_forms_import_job(
 $1,$2,$3,false,$4)`, call.tenantID, call.importID, result.jobID,
		call.createdAt).Scan(&state, &cancelled, &completedAt, &changed)
	if err != nil || state != "queued" || cancelled || changed || completedAt != nil {
		t.Fatalf("retry control = %s %t %t %v, %v",
			state, cancelled, changed, completedAt, err)
	}
	attemptID, leaseOwnerID := mustRetryV7(t), uuid.New()
	startedAt := time.Now().UTC()
	_, err = database.ExecContext(ctx, `
UPDATE platform.jobs SET state='running',attempts=1,lease_generation=1,
 lease_owner_id=$1,lease_heartbeat_at=$2,lease_until=$3,updated_at=$2
WHERE id=$4`, leaseOwnerID, startedAt,
		startedAt.Add(time.Minute), result.jobID)
	if err != nil {
		t.Fatalf("prepare retry job lease: %v", err)
	}
	_, err = database.ExecContext(ctx, `
INSERT INTO platform.job_attempts(
 attempt_id,job_id,scope_kind,scope_id,attempt,lease_generation,
 lease_owner_id,started_at
) VALUES($1,$2,'tenant',$3,1,1,$4,$5)`, attemptID, result.jobID,
		call.tenantID, leaseOwnerID, startedAt)
	if err != nil {
		t.Fatalf("prepare retry attempt lease: %v", err)
	}
	var current bool
	err = runtime.QueryRow(ctx, `SELECT public.lease_forms_import_job(
 $1,$2,$3,$4,$5,1)`, call.tenantID, call.importID, result.jobID,
		attemptID, leaseOwnerID).Scan(&current)
	if err != nil || current {
		t.Fatalf("retry lease = %t, %v", current, err)
	}
	var completed bool
	err = runtime.QueryRow(ctx, `SELECT public.complete_forms_import_job(
 $1,$2,$3,$4,$5,1,$6)`, call.tenantID, call.importID, result.jobID,
		attemptID, leaseOwnerID, startedAt.Add(time.Second)).Scan(&completed)
	if err != nil || !completed {
		t.Fatalf("retry completion = %t, %v", completed, err)
	}
	var attemptOutcome string
	var jobFinished, attemptFinished bool
	err = database.QueryRowContext(ctx, `
SELECT job.state,job.completed_at IS NOT NULL,attempt.outcome,
       attempt.finished_at IS NOT NULL
FROM platform.jobs AS job JOIN platform.job_attempts AS attempt
 ON attempt.job_id=job.id WHERE job.id=$1`, result.jobID).Scan(
		&state, &jobFinished, &attemptOutcome, &attemptFinished)
	if err != nil || state != "succeeded" || !jobFinished ||
		attemptOutcome != "succeeded" || !attemptFinished {
		t.Fatalf("retry completion rows = %s %t %s %t, %v",
			state, jobFinished, attemptOutcome, attemptFinished, err)
	}
}
