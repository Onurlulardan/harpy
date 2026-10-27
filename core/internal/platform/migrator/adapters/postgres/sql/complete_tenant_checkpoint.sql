UPDATE platform.tenant_migration_checkpoints
SET state = $9, finished_at = pg_catalog.clock_timestamp(),
    error_code = $10, updated_at = pg_catalog.clock_timestamp()
WHERE tenant_id = $1 AND owner = $2 AND logical_version = $3
  AND attempt = $4 AND database_origin = $5 AND selected_path = $6
  AND selected_sha256 = $7 AND run_id = $8 AND state = 'running';
