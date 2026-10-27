UPDATE platform.tenant_migration_checkpoints
SET run_id = $4, state = 'running', attempt = $5,
    database_origin = $6, selected_path = $7, selected_sha256 = $8
WHERE tenant_id = $1 AND owner = $2 AND logical_version = $3
  AND database_origin = $9 AND selected_path = $10
  AND selected_sha256 = $11
  AND state = $12 AND attempt = $13 AND run_id = $14;
