INSERT INTO catalog_menu_items (
    version_id, item_id, parent_item_id, package_id, item_kind,
    target_id, external_url, label, sort_order
)
SELECT $1::uuid, item_id, parent_item_id, package_id, item_kind,
       target_id, external_url, label, sort_order
FROM catalog_menu_items
WHERE version_id = $2::uuid AND package_id <> $3::uuid
