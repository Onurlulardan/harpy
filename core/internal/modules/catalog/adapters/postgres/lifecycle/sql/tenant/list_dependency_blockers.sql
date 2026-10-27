SELECT dependent.package_id::text, dependent.active_version, resolution.required
FROM catalog_dependency_resolutions AS resolution
JOIN catalog_installations AS dependent
  ON dependent.installation_id = resolution.installation_id
WHERE resolution.dependency_package_id = $1::uuid
  AND dependent.state IN ('ACTIVE', 'DISABLED')
ORDER BY dependent.package_id
