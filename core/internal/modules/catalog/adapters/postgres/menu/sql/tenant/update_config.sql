UPDATE catalog_menu_configs
SET name = $2::text,
    config = $3::jsonb,
    is_active = $4::boolean,
    updated_by = $5::uuid,
    updated_at = $6::timestamptz
WHERE config_id = $1::uuid
RETURNING config_id::text, name, config, is_active,
    created_by::text, updated_by::text, created_at, updated_at
