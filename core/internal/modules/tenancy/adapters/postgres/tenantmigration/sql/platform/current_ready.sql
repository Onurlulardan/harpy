SELECT COALESCE(ready.release_id::text, ''),
       COALESCE(ready.environment_id::text, ''),
       COALESCE(ready.deployment_mode, ''),
       COALESCE(ready.runtime_generation, 0),
       COALESCE(ready.database_generation, 0),
       COALESCE(ready.migration_manifest_sha256, ''),
       COALESCE(ready.ready_until, 'epoch'::timestamptz),
       ready.release_id IS NOT NULL AS found
FROM (VALUES (true)) AS sentinel(present)
LEFT JOIN LATERAL (
    SELECT *
    FROM platform.tenancy_tenant_migration_ready_current_v1()
    LIMIT 1
) AS ready ON sentinel.present
