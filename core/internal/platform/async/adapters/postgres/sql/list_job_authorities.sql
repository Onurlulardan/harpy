WITH job_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS observed_at
), authorities AS MATERIALIZED (
    SELECT group_row.scope_kind,group_row.scope_id,group_row.group_id,
           group_row.generation AS group_generation,domain_row.domain,
           domain_row.generation,'v2'::text AS mode
    FROM platform.cutover_group_members AS member
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind=member.scope_kind
     AND group_row.scope_id=member.scope_id
     AND group_row.group_id=member.group_id
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind=member.scope_kind
     AND domain_row.scope_id=member.scope_id
     AND domain_row.domain=member.domain
    CROSS JOIN job_clock
    WHERE group_row.state='v2' AND group_row.writer='v2'
      AND domain_row.state='v2' AND domain_row.writer='v2'
      AND EXISTS (
          SELECT 1 FROM platform.jobs AS job
          WHERE job.scope_kind=group_row.scope_kind
            AND job.scope_id=group_row.scope_id
            AND job.authority_group_id=group_row.group_id
            AND job.authority_group_generation=group_row.generation
            AND job.authority_domain=domain_row.domain
            AND job.cutover_generation=domain_row.generation
            AND job.parked_at IS NULL
            AND NOT EXISTS (
                SELECT 1 FROM platform.cutover_job_bindings AS binding
                WHERE binding.job_id=job.id
            )
            AND ((job.state='queued' AND job.attempts<job.max_attempts
                  AND job.run_at<=job_clock.observed_at)
              OR (job.state='running' AND job.lease_until IS NOT NULL
                  AND job.lease_until<=job_clock.observed_at))
      )
    UNION ALL
    SELECT group_row.scope_kind,group_row.scope_id,group_row.group_id,
           group_row.generation,domain_row.domain,domain_row.generation,
           'cutover_shadow'::text AS mode
    FROM platform.cutover_group_members AS member
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind=member.scope_kind
     AND group_row.scope_id=member.scope_id
     AND group_row.group_id=member.group_id
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind=member.scope_kind
     AND domain_row.scope_id=member.scope_id
     AND domain_row.domain=member.domain
    CROSS JOIN job_clock
    WHERE group_row.scope_kind='tenant'
      AND group_row.group_id='documents'
      AND group_row.state IN ('v1','shadow') AND group_row.writer='v1'
      AND domain_row.domain='documents'
      AND domain_row.state IN ('v1','shadow') AND domain_row.writer='v1'
      AND EXISTS (
          SELECT 1 FROM platform.jobs AS job
          JOIN platform.cutover_job_bindings AS binding ON binding.job_id=job.id
          WHERE job.scope_kind=group_row.scope_kind
            AND job.scope_id=group_row.scope_id
            AND job.authority_group_id=group_row.group_id
            AND job.authority_group_generation=group_row.generation
            AND job.authority_domain=domain_row.domain
            AND job.cutover_generation=domain_row.generation
            AND job.writer_origin='v2' AND job.kind='documents.adoption.copy'
            AND job.parked_at IS NULL
            AND binding.tenant_id=job.scope_id
            AND binding.group_generation=job.authority_group_generation
            AND binding.cutover_generation=job.cutover_generation
            AND binding.payload_ref=job.payload_ref
            AND binding.payload_sha256=job.payload_sha256
            AND platform.platform_async_cutover_shadow_operation_active_v1(
                binding.operation_id,binding.cutover_run_id,
                job.scope_kind,job.scope_id,job.authority_group_id,
                job.authority_domain,job.cutover_generation,
                binding.lease_manifest_sha256,job_clock.observed_at
            )
            AND ((job.state='queued' AND job.attempts<job.max_attempts
                  AND job.run_at<=job_clock.observed_at)
              OR (job.state='running' AND job.lease_until IS NOT NULL
                  AND job.lease_until<=job_clock.observed_at))
      )
)
SELECT scope_kind,scope_id,group_id,group_generation,domain,generation,mode
FROM authorities AS authority
WHERE NOT $1::boolean OR ROW(
    authority.scope_kind COLLATE "C",authority.scope_id,
    authority.group_id COLLATE "C",authority.domain COLLATE "C"
) > ROW(
    $2::text COLLATE "C",$3::uuid,$4::text COLLATE "C",$5::text COLLATE "C"
)
ORDER BY authority.scope_kind COLLATE "C",authority.scope_id,
         authority.group_id COLLATE "C",authority.domain COLLATE "C"
LIMIT $6;
