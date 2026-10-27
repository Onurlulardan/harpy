INSERT INTO platform.tenant_migration_checkpoints (
    tenant_id, owner, logical_version, database_origin,
    selected_path, selected_sha256, run_id, state, attempt
) VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', 1);
