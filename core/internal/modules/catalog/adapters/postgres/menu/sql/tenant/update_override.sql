UPDATE catalog_menu_overrides
SET label = COALESCE($2::text, label),
    sort_order = COALESCE($3::integer, sort_order),
    visible = COALESCE($4::boolean, visible),
    revision = revision + 1
WHERE item_id = $1::uuid AND revision = $5::bigint
RETURNING item_id::text, label, sort_order, visible, revision
