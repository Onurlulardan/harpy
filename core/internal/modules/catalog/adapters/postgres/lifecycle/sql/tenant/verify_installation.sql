SELECT installation_id
FROM catalog_installations
WHERE installation_id = $1::uuid AND package_id = $2::uuid AND row_version = $3::bigint
