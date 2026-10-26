WITH selected_definition AS (
  SELECT version.definition_id
  FROM catalog_menu_versions AS version
  JOIN catalog_menu_items AS item ON item.version_id=version.version_id
  WHERE item.item_id=$1
)
SELECT version.version_number,version.created_at,
       definition.active_version_id=version.version_id
FROM catalog_menu_versions AS version
JOIN catalog_menu_definitions AS definition
  ON definition.definition_id=version.definition_id
WHERE version.definition_id=(SELECT definition_id FROM selected_definition)
ORDER BY version.version_number DESC
LIMIT $2
