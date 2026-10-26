SELECT package_id::text
FROM catalog_installations
WHERE state = 'ACTIVE'
ORDER BY package_id
