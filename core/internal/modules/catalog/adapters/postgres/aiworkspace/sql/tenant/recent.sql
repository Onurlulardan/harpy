SELECT item.item_id,item.package_id,item.target_id,item.item_kind,
       COALESCE(override.label,item.label),version.version_number,
       version.created_at,0::double precision
FROM catalog_menu_definitions AS definition
JOIN catalog_menu_versions AS version ON version.version_id=definition.active_version_id
JOIN catalog_menu_items AS item ON item.version_id=version.version_id
JOIN catalog_installations AS installation ON installation.package_id=item.package_id
LEFT JOIN catalog_menu_overrides AS override ON override.item_id=item.item_id
WHERE installation.state='ACTIVE' AND COALESCE(override.visible,true)
  AND item.item_kind<>'EXTERNAL_LINK' AND version.created_at>$1
  AND (pg_catalog.cardinality($2::text[])=0 OR item.item_kind=ANY($2::text[]))
ORDER BY version.created_at DESC,item.item_id
LIMIT $3
