SELECT version.package_id::text, package.slug, version.version,
       version.manifest, version.manifest_sha256, version.signature,
       version.key_id, version.release_id::text, version.registered_at
FROM platform.catalog_package_versions AS version
JOIN platform.catalog_packages AS package ON package.package_id = version.package_id
WHERE version.package_id = $1::uuid
ORDER BY version.registered_at DESC, version.version DESC
LIMIT $2::integer
