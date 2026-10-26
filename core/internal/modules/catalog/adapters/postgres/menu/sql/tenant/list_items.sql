SELECT item_id::text, parent_item_id::text, package_id::text, item_kind,
       target_id::text, external_url, label, sort_order
FROM catalog_menu_items
WHERE version_id = $1::uuid
ORDER BY sort_order, item_id
