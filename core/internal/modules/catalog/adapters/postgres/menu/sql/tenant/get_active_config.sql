SELECT config.config_id::text, config.name, config.config, config.is_active,
    config.created_by::text, config.updated_by::text, config.created_at, config.updated_at
FROM catalog_menu_configs AS config
WHERE config.is_active
LIMIT 1
