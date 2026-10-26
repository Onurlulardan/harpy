WITH root AS (
    INSERT INTO notifications_provider_configs (
        id, provider_kind, name, enabled, current_version_id,
        row_version, created_at, updated_at
    ) VALUES (
        $1::uuid, $4, $5, $6, NULL, $7,
        pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp()
    )
    RETURNING id
), version AS (
    INSERT INTO notifications_provider_config_versions (
        id, config_id, version_number, secret_id, metadata,
        metadata_digest, created_at
    )
    SELECT $2::uuid, id, $7, $3::uuid, $8::jsonb, $9,
           pg_catalog.transaction_timestamp()
    FROM root
    RETURNING config_id, id
)
UPDATE notifications_provider_configs config
SET current_version_id = version.id
FROM version
WHERE config.id = version.config_id;
