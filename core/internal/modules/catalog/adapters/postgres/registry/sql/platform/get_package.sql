SELECT package_id::text, slug, created_at
FROM platform.catalog_packages
WHERE package_id = $1::uuid
