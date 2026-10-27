WITH completed AS (
    UPDATE platform.jobs AS job
    SET state = 'succeeded',
        lease_owner_id = NULL,
        lease_heartbeat_at = NULL,
        lease_until = NULL,
        last_error_class = NULL,
        last_error_code = NULL,
        last_error_sha256 = NULL,
        retain_until = $11::timestamptz + interval '90 days',
        updated_at = $11::timestamptz,
        completed_at = $11::timestamptz
    WHERE job.scope_kind = $1
      AND job.scope_id = $2
      AND job.authority_group_id = $3
      AND job.authority_group_generation = $4
      AND job.authority_domain = $5
      AND job.cutover_generation = $6
      AND job.id = $7
      AND job.state = 'running'
      AND job.effect_class = 'external'
      AND job.effect_started_at = $10::timestamptz
      AND job.lease_owner_id = $8
      AND job.lease_generation = $9
    RETURNING job.completed_at
)
SELECT completed_at
FROM completed;
