SELECT version,
       record_kind,
       is_applied,
       owner,
       scope,
       origin,
       selected_path,
       selected_sha256,
       run_id::text,
       target_database_generation,
       next_minimum_runtime_generation
FROM platform.platform_migration_history
WHERE version > 0
ORDER BY version;
