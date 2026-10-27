WITH authoritative_time AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
INSERT INTO platform.schema_job_step_attempts (
    job_id,
    step_ordinal,
    attempt,
    plan_sha256,
    step_sha256,
    lease_id,
    lease_owner_id,
    executor_kind,
    state,
    started_at,
    finished_at,
    outcome_sha256,
    error_code,
    redacted_detail_sha256,
    recorded_at
)
SELECT
    $1, $2, $3, $4, $5, $6, $7, 'migrator', $8,
    $9, authoritative_time.now, $10, $11, $12, authoritative_time.now
FROM authoritative_time
RETURNING finished_at;
