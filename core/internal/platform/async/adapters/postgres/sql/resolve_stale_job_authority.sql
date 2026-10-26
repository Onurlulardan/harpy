SELECT
    authority_group.scope_kind,
    authority_group.scope_id,
    authority_group.group_id,
    authority_group.generation,
    authority_domain.domain,
    authority_domain.generation
FROM platform.jobs AS job
JOIN platform.cutover_groups AS authority_group
  ON authority_group.scope_kind = job.scope_kind
 AND authority_group.scope_id = job.scope_id
 AND authority_group.group_id = job.authority_group_id
JOIN platform.domain_cutovers AS authority_domain
  ON authority_domain.scope_kind = job.scope_kind
 AND authority_domain.scope_id = job.scope_id
 AND authority_domain.domain = job.authority_domain
JOIN platform.cutover_group_members AS member
  ON member.scope_kind = job.scope_kind
 AND member.scope_id = job.scope_id
 AND member.group_id = job.authority_group_id
 AND member.domain = job.authority_domain
WHERE job.id = $1::uuid
  AND job.state IN ('queued', 'cancelled')
  AND job.parked_at IS NOT NULL
  AND job.parked_code = 'JOB_AUTHORITY_STALE'
  AND job.effect_started_at IS NULL
  AND authority_group.state = 'v2'
  AND authority_group.writer = 'v2'
  AND authority_domain.state = 'v2'
  AND authority_domain.writer = 'v2'
  AND authority_group.generation >= job.parked_current_group_generation
  AND authority_domain.generation >= job.parked_current_cutover_generation
  AND authority_group.generation >= job.authority_group_generation
  AND authority_domain.generation >= job.cutover_generation
  AND (
      authority_group.generation > job.authority_group_generation OR
      authority_domain.generation > job.cutover_generation
  );
