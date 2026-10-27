WITH candidates AS (
    SELECT *
    FROM platform.jobs
    WHERE id = $1 OR (
        scope_kind = 'tenant' AND scope_id = $2 AND
        authority_group_id = $3 AND authority_domain = $4 AND
        kind = $5 AND dedupe_key = $6
    )
)
SELECT
    COALESCE(pg_catalog.min(id::text), ''),
    pg_catalog.count(*) = 1 AND COALESCE(pg_catalog.bool_and(
        tenant_id = $2 AND scope_kind = 'tenant' AND scope_id = $2 AND
        authority_group_id = $3 AND authority_group_generation = $7 AND
        authority_domain = $4 AND cutover_generation = $8 AND writer_origin = 'v2' AND
        kind = $5 AND dedupe_key = $6 AND payload_revision = $9 AND
        payload_sha256 = $10 AND effect_class = $11 AND
        effect_idempotency_key IS NOT DISTINCT FROM $12 AND priority = $13 AND
        max_attempts = $14 AND retry_of_job_id IS NOT DISTINCT FROM $15 AND
        run_at = $16 AND (
            payload_ref = $17 OR
            (payload_ref IS NULL AND payload_redacted_at IS NOT NULL)
        )
    ), false)
FROM candidates;
