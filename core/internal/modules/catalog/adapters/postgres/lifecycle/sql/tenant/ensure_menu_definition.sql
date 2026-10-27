INSERT INTO catalog_menu_definitions (
    definition_id, revision, active_version_id, updated_at
) VALUES ($1::uuid, 1, NULL, $2::timestamptz)
ON CONFLICT (definition_id) DO NOTHING
