SELECT run_id::text,
       release_id::text,
       environment_id::text,
       database_origin,
       command,
       state,
       migration_manifest_sha256,
       extension_manifest_sha256,
       roles_manifest_sha256,
       privileges_manifest_sha256,
       ownership_manifest_sha256,
       COALESCE(failure_code, '')
FROM platform.migration_runs
WHERE run_id = $1;
