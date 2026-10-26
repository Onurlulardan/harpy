WITH current AS (
    SELECT id FROM notifications_provider_configs
    WHERE id = $1::uuid AND row_version = $7
    FOR UPDATE
), secret AS (
    INSERT INTO notifications_provider_secrets (
        id, ciphertext, nonce, wrapped_dek, kek_revision, plaintext_digest,
        created_at, rewrapped_at
    )
    SELECT $3::uuid, $9, $10, $11, $12, $13,
           pg_catalog.transaction_timestamp(), NULL
    FROM current
    RETURNING id
), version AS (
    INSERT INTO notifications_provider_config_versions (
        id, config_id, version_number, secret_id, metadata,
        metadata_digest, created_at
    )
    SELECT $2::uuid, current.id, $8, secret.id, $14::jsonb, $15,
           pg_catalog.transaction_timestamp()
    FROM current CROSS JOIN secret
    RETURNING config_id, id
)
UPDATE notifications_provider_configs config
SET current_version_id = version.id, row_version = $8,
    updated_at = pg_catalog.transaction_timestamp()
FROM version
WHERE config.id = version.config_id;
