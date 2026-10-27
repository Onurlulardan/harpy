SELECT override.item_id::text, override.label, override.sort_order, override.visible
FROM catalog_menu_overrides AS override
JOIN catalog_menu_items AS item ON item.item_id = override.item_id
WHERE item.version_id = $1::uuid
ORDER BY override.item_id
