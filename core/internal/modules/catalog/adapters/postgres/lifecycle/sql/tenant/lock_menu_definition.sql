SELECT revision, active_version_id
FROM catalog_menu_definitions
WHERE definition_id = $1::uuid
FOR UPDATE
