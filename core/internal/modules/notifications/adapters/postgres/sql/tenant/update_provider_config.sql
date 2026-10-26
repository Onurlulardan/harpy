WITH current AS (
    SELECT id FROM notifications_provider_configs
    WHERE id = $1::uuid AND row_version = $7
    FOR UPDATE
), version AS (
    INSERT INTO notifications_provider_config_versions (
        id, config_id, version_number, secret_id, metadata,
        metadata_digest, created_at
    )
    SELECT $2::uuid, current.id, $8, $3::uuid, $9::jsonb, $10,
           pg_catalog.transaction_timestamp()
    FROM current
    RETURNING config_id, id
)
UPDATE notifications_provider_configs config
SET name = $5, enabled = $6, current_version_id = version.id,
    row_version = $8, updated_at = pg_catalog.transaction_timestamp()
FROM version
WHERE config.id = version.config_id AND config.provider_kind = $4;
