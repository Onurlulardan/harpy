INSERT INTO catalog_menu_overrides (
    item_id, label, sort_order, visible, revision
) VALUES (
    $1::uuid, NULL, NULL, false, 1
)
ON CONFLICT (item_id) DO UPDATE
SET visible = false;
