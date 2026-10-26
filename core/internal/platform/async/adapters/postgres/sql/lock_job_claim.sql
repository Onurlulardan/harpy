SELECT
    job.cancel_requested_at,
    job.effect_started_at,
    job.lease_until
FROM platform.jobs AS job
WHERE job.scope_kind = $1
  AND job.scope_id = $2
  AND job.authority_group_id = $3
  AND job.authority_group_generation = $4
  AND job.authority_domain = $5
  AND job.cutover_generation = $6
  AND job.id = $7
  AND job.state = 'running'
  AND job.lease_owner_id = $8
  AND job.lease_generation = $9
  AND job.lease_until > pg_catalog.clock_timestamp()
FOR UPDATE;
