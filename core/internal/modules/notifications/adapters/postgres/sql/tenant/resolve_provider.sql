SELECT config.id::text, version.id::text, version.secret_id::text,
       config.provider_kind, config.name, config.enabled, config.row_version,
       version.metadata, version.metadata_digest
FROM notifications_provider_configs config
JOIN notifications_provider_config_versions version
  ON version.id = COALESCE(NULLIF($2, '')::uuid, config.current_version_id)
WHERE config.provider_kind = CASE $1 WHEN 'email' THEN 'smtp' ELSE 'telegram' END
  AND version.config_id = config.id
  AND (NULLIF($2, '')::uuid IS NOT NULL OR config.enabled)
LIMIT 1;
