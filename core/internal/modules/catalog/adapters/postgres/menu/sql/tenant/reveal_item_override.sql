UPDATE catalog_menu_overrides
SET visible = true
WHERE item_id = $1::uuid;
