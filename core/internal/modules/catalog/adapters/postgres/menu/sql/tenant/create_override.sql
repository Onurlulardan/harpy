INSERT INTO catalog_menu_overrides (
    item_id, label, sort_order, visible, revision
) VALUES (
    $1::uuid, $2::text, $3::integer, $4::boolean, 1
)
ON CONFLICT DO NOTHING
RETURNING item_id::text, label, sort_order, visible, revision
