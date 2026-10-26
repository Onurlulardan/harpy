UPDATE catalog_menu_configs
SET is_active = false,
    updated_by = $2::uuid,
    updated_at = $3::timestamptz
WHERE is_active AND config_id <> $1::uuid
