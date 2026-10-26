SELECT config.id::text, version.id::text, version.secret_id::text,
       config.provider_kind, config.name, config.enabled, config.row_version,
       version.metadata, version.metadata_digest
FROM notifications_provider_config_versions version
JOIN notifications_provider_configs config ON config.id = version.config_id
WHERE version.id = $1::uuid;
