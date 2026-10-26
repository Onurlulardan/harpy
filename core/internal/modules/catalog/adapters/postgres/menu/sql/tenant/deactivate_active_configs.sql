UPDATE catalog_menu_configs
SET is_active = false,
    updated_by = $1::uuid,
    updated_at = $2::timestamptz
WHERE is_active
