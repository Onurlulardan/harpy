SELECT package_id::text, slug, created_at
FROM platform.catalog_packages
WHERE slug = $1::text
