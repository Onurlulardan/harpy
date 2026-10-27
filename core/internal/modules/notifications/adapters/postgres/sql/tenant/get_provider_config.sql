SELECT config.id::text, version.id::text, version.secret_id::text,
       config.provider_kind, config.name, config.enabled, config.row_version,
       version.metadata, version.metadata_digest
FROM notifications_provider_configs config
JOIN notifications_provider_config_versions version ON version.id = config.current_version_id
WHERE config.id = $1::uuid;
