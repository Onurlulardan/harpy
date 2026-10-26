SELECT version.version_id::text, version.version_number, definition.revision
FROM catalog_menu_definitions AS definition
JOIN catalog_menu_versions AS version
  ON version.version_id = definition.active_version_id
WHERE definition.definition_id = $1::uuid
