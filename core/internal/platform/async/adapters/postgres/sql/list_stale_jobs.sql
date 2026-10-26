SELECT
    job.id,
    job.state,
    job.lease_owner_id,
    job.lease_generation,
    job.kind,
    job.payload_revision,
    job.payload_ref,
    job.payload_sha256
FROM platform.jobs AS job
JOIN platform.cutover_groups AS group_row
  ON group_row.scope_kind = job.scope_kind
 AND group_row.scope_id = job.scope_id
 AND group_row.group_id = job.authority_group_id
JOIN platform.cutover_group_members AS member
  ON member.scope_kind = job.scope_kind
 AND member.scope_id = job.scope_id
 AND member.group_id = job.authority_group_id
 AND member.domain = job.authority_domain
JOIN platform.domain_cutovers AS domain_row
  ON domain_row.scope_kind = job.scope_kind
 AND domain_row.scope_id = job.scope_id
 AND domain_row.domain = job.authority_domain
WHERE group_row.state = 'v2'
  AND group_row.writer = 'v2'
  AND domain_row.state = 'v2'
  AND domain_row.writer = 'v2'
  AND group_row.generation >= job.authority_group_generation
  AND domain_row.generation >= job.cutover_generation
  AND (group_row.generation > job.authority_group_generation OR
       domain_row.generation > job.cutover_generation)
  AND job.state IN ('queued', 'running')
  AND job.parked_at IS NULL
  AND job.effect_started_at IS NULL
ORDER BY job.created_at, job.id
LIMIT $1;

