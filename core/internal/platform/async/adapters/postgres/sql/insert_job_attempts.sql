INSERT INTO platform.job_attempts (
    attempt_id,
    job_id,
    scope_kind,
    scope_id,
    attempt,
    lease_generation,
    lease_owner_id,
    started_at
)
SELECT
    pg_catalog.uuidv7(),
    ($1::uuid[])[input.position],
    ($2::text[])[input.position],
    ($3::uuid[])[input.position],
    ($4::integer[])[input.position],
    ($5::bigint[])[input.position],
    ($6::uuid[])[input.position],
    ($7::timestamptz[])[input.position]
FROM pg_catalog.generate_subscripts($1::uuid[], 1) AS input(position)
RETURNING attempt_id, job_id;
