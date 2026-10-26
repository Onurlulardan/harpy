INSERT INTO catalog_menu_items (
    version_id, item_id, parent_item_id, package_id, item_kind,
    target_id, external_url, label, sort_order
) VALUES (
    $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::text,
    $6::uuid, $7::text, $8::text, $9::integer
)
