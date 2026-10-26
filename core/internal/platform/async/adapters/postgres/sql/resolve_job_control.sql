SELECT job.id
FROM platform.jobs AS job
WHERE job.scope_kind = $1
  AND job.scope_id = $2
  AND job.authority_group_id = $3
  AND job.authority_group_generation = $4
  AND job.authority_domain = $5
  AND job.cutover_generation = $6
  AND job.id = $7
  AND job.state IN ('queued', 'running');
