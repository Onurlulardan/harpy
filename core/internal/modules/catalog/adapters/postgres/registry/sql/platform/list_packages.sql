SELECT package_id::text, slug, created_at
FROM platform.catalog_packages
WHERE $1::uuid IS NULL OR package_id > $1::uuid
ORDER BY package_id
LIMIT $2::integer
