//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertSchemaTransitionTriggerState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	expected := map[string]string{
		"schema_job_step_attempts_validate_insert":    "D",
		"schema_job_step_attempts_validate_insert_v2": "O",
		"schema_job_steps_validate_transition":        "D",
		"schema_job_steps_validate_transition_v2":     "D",
		"schema_job_steps_validate_transition_v3":     "O",
		"schema_jobs_validate_transition":             "D",
		"schema_jobs_validate_transition_v2":          "O",
		"schema_jobs_validate_revision":               "O",
	}
	rows, err := database.QueryContext(ctx, `SELECT trigger_row.tgname,
trigger_row.tgenabled::text FROM pg_catalog.pg_trigger AS trigger_row
WHERE trigger_row.tgname=ANY($1)`, triggerNames(expected))
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := map[string]string{}
	for rows.Next() {
		var name, state string
		if err := rows.Scan(&name, &state); err != nil {
			t.Fatal(err)
		}
		seen[name] = state
	}
	for name, state := range expected {
		if seen[name] != state {
			t.Fatalf("trigger %s state = %q, want %q", name, seen[name], state)
		}
	}
}

func triggerNames(values map[string]string) []string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	return names
}

func assertBackfillProgressAccepted(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `UPDATE platform.schema_job_steps SET
generation=generation+1,checkpoint_record_id=$2,processed_rows=processed_rows+100,
updated_at=pg_catalog.clock_timestamp() WHERE job_id=$1 AND ordinal=0`,
		progressJobID, "018f0000-0000-7000-8000-000000000711")
	if err != nil {
		t.Fatalf("running backfill progress rejected: %v", err)
	}
	var rows int64
	if err := database.QueryRowContext(ctx, `SELECT processed_rows
FROM platform.schema_job_steps WHERE job_id=$1 AND ordinal=0`, progressJobID).Scan(&rows); err != nil || rows != 100 {
		t.Fatalf("backfill progress = %d err=%v", rows, err)
	}
}

func assertMultiStepAttemptDivergence(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	finished := time.Now().UTC().Add(-time.Millisecond).Truncate(time.Microsecond)
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, schemaAttemptInsertSQL,
		multiStepJobID, 1, finished, "succeeded",
		integrationDigest("multi-outcome"), nil, nil); err != nil {
		t.Fatalf("multi-step receipt with divergent job attempt rejected: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE platform.schema_job_steps SET
generation=generation+1,state='succeeded',next_executor='none',finished_at=$2,
updated_at=$2 WHERE job_id=$1 AND ordinal=1`, multiStepJobID, finished); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE platform.schema_jobs SET
generation=generation+1,state='succeeded',next_step_ordinal=2,next_executor='none',
lease_id=NULL,lease_owner_id=NULL,leased_at=NULL,lease_heartbeat_at=NULL,
lease_expires_at=NULL,finished_at=$2,updated_at=$2
WHERE job_id=$1`, multiStepJobID, finished); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("multi-step jobAttempt != stepAttempt rejected: %v", err)
	}
}

const schemaAttemptInsertSQL = `INSERT INTO platform.schema_job_step_attempts
(job_id,step_ordinal,attempt,plan_sha256,step_sha256,lease_id,lease_owner_id,
executor_kind,state,started_at,finished_at,outcome_sha256,error_code,
redacted_detail_sha256)
SELECT step.job_id,step.ordinal,step.attempt,step.plan_sha256,step.step_sha256,
job.lease_id,job.lease_owner_id,step.executor_kind,$4,step.started_at,$3,$5,$6,$7
FROM platform.schema_job_steps AS step
JOIN platform.schema_jobs AS job ON job.job_id=step.job_id
WHERE step.job_id=$1 AND step.ordinal=$2`
