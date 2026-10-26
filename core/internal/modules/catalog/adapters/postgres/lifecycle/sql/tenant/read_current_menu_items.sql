SELECT item.item_id::text, item.parent_item_id::text, item.package_id::text,
       item.item_kind, item.target_id::text, item.external_url,
       item.label, item.sort_order
FROM catalog_menu_definitions AS definition
JOIN catalog_menu_items AS item ON item.version_id = definition.active_version_id
WHERE definition.definition_id = $1::uuid
ORDER BY item.item_id
