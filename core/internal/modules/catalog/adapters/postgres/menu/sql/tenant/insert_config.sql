INSERT INTO catalog_menu_configs (
    config_id, name, config, is_active, created_by, updated_by, created_at, updated_at
) VALUES (
    $1::uuid, $2::text, $3::jsonb, true, $4::uuid, $4::uuid, $5::timestamptz, $5::timestamptz
)
RETURNING config_id::text, name, config, is_active,
    created_by::text, updated_by::text, created_at, updated_at
