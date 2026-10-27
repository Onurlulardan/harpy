SELECT job.id
FROM platform.jobs AS job
WHERE job.id = $1::uuid
  AND job.scope_kind = 'tenant'
  AND job.scope_id = $2::uuid
  AND job.authority_group_id = $3
  AND job.authority_group_generation = $4
  AND job.authority_domain = $5
  AND job.cutover_generation = $6
  AND job.writer_origin = 'v2'
  AND job.state = 'running'
  AND job.effect_started_at IS NOT NULL
  AND job.cancel_requested_at IS NULL;
