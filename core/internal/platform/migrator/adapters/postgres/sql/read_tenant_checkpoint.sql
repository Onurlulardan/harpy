SELECT database_origin, selected_path, selected_sha256, run_id::text,
       state, attempt, started_at, finished_at, error_code, updated_at
FROM platform.tenant_migration_checkpoints
WHERE tenant_id = $1 AND owner = $2 AND logical_version = $3;
