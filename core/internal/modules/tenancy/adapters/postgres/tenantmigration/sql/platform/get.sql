SELECT COALESCE(job.job_id::text, ''), COALESCE(job.operation_id::text, ''),
       COALESCE(job.state, ''), COALESCE(job.tenant_id::text, ''),
       COALESCE(job.schema_name, ''), COALESCE(job.step_plan_sha256, ''),
       COALESCE(job.release_id::text, ''), COALESCE(job.runtime_generation, 0),
       COALESCE(job.database_generation, 0),
       COALESCE(job.migration_manifest_sha256, ''),
       job.error_code, job.completion_receipt_id::text,
       job.completion_receipt_sha256, job.completion_evidence_sha256,
       job.completed_at, job.created_at, job.database_now,
       job.job_id IS NOT NULL AS found
FROM (VALUES (true)) AS sentinel(present)
LEFT JOIN LATERAL (
    SELECT *
    FROM platform.tenancy_tenant_migration_get_v2($1)
    LIMIT 1
) AS job ON sentinel.present
