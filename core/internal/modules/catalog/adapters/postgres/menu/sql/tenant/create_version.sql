INSERT INTO catalog_menu_versions (
    version_id, definition_id, version_number, content_sha256, created_at
) VALUES ($1::uuid, $2::uuid, $3::bigint, $4::text, $5::timestamptz)
