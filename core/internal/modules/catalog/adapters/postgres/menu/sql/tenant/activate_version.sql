UPDATE catalog_menu_definitions
SET active_version_id = $2::uuid,
    revision = $3::bigint,
    updated_at = $4::timestamptz
WHERE definition_id = $1::uuid AND revision = $5::bigint
