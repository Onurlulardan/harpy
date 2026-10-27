//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"testing"
)

func seedSchemaStates(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	seedFutureSchemaState(t, ctx, database, progressJobID, 0, 2, 1)
	seedFutureSchemaState(t, ctx, database, multiStepJobID, 1, 3, 1)
	seedFutureSchemaState(t, ctx, database, permanentFailureJobID, 0, 1, 1)
	seedFutureSchemaState(t, ctx, database, activeAttemptEightJobID, 0, 8, 8)
	_, err := database.ExecContext(ctx, `UPDATE platform.schema_job_steps SET
generation=2,state='succeeded',next_executor='none',attempt=1,
started_at=pg_catalog.clock_timestamp(),finished_at=pg_catalog.clock_timestamp(),
updated_at=pg_catalog.clock_timestamp()
WHERE job_id=$1 AND ordinal=0`, multiStepJobID)
	if err != nil {
		t.Fatal(err)
	}
	seedExpiredSchemaState(t, ctx, database, expiredLeasedJobID, "leased")
	seedExpiredSchemaState(t, ctx, database, expiredRunningJobID, "running")
}

func seedFutureSchemaState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	jobID string,
	ordinal int,
	jobAttempt int,
	stepAttempt int,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `UPDATE platform.schema_jobs SET
generation=4,state='running',next_step_ordinal=$2,attempt=$3,
lease_id='018f0000-0000-7000-8000-000000000790',
lease_owner_id='018f0000-0000-7000-8000-000000000791',
leased_at=pg_catalog.clock_timestamp(),
lease_heartbeat_at=pg_catalog.clock_timestamp(),
lease_expires_at=pg_catalog.clock_timestamp()+interval '5 minutes',
updated_at=pg_catalog.clock_timestamp() WHERE job_id=$1`, jobID, ordinal, jobAttempt)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.ExecContext(ctx, `UPDATE platform.schema_job_steps SET
generation=3,state='running',attempt=$3,started_at=pg_catalog.clock_timestamp(),
updated_at=pg_catalog.clock_timestamp() WHERE job_id=$1 AND ordinal=$2`,
		jobID, ordinal, stepAttempt)
	if err != nil {
		t.Fatal(err)
	}
}

func seedExpiredSchemaState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	jobID string,
	state string,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `WITH moment AS (
SELECT pg_catalog.clock_timestamp()-interval '2 minutes' AS value)
UPDATE platform.schema_jobs SET generation=4,state=$2,attempt=8,
lease_id='018f0000-0000-7000-8000-000000000792',
lease_owner_id='018f0000-0000-7000-8000-000000000793',
leased_at=moment.value,lease_heartbeat_at=moment.value,
lease_expires_at=moment.value+interval '1 minute',
created_at=moment.value,updated_at=moment.value FROM moment WHERE job_id=$1`, jobID, state)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.ExecContext(ctx, `WITH moment AS (
SELECT pg_catalog.clock_timestamp()-interval '2 minutes' AS value)
UPDATE platform.schema_job_steps SET generation=3,state='running',attempt=8,
started_at=moment.value,created_at=moment.value,updated_at=moment.value
FROM moment WHERE job_id=$1 AND ordinal=0`, jobID)
	if err != nil {
		t.Fatal(err)
	}
}

func seedExpiredAttemptHistory(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	seed schemaSeed,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `INSERT INTO platform.schema_job_step_attempts
(job_id,step_ordinal,attempt,plan_sha256,step_sha256,lease_id,lease_owner_id,
executor_kind,state,started_at,finished_at,outcome_sha256,error_code,
redacted_detail_sha256)
SELECT $1,0,value,$2,$3,pg_catalog.gen_random_uuid(),
'018f0000-0000-7000-8000-000000000793','migrator','lease_expired',
pg_catalog.clock_timestamp()-interval '10 minutes',
pg_catalog.clock_timestamp()-interval '9 minutes',$4,'lease_expired',$5
FROM pg_catalog.generate_series(1,7) AS value`, seed.jobID, seed.planHash,
		seed.stepHash, integrationDigest("expired-outcome"), integrationDigest("expired-detail"))
	if err != nil {
		t.Fatal(err)
	}
}
