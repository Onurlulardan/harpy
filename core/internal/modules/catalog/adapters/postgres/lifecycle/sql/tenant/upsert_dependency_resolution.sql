INSERT INTO catalog_dependency_resolutions (
    installation_id, dependency_package_id, exact_version, required, plan_sha256
) VALUES ($1::uuid, $2::uuid, $3::text, $4::boolean, $5::text)
ON CONFLICT (installation_id, dependency_package_id) DO UPDATE SET
    exact_version = EXCLUDED.exact_version,
    required = EXCLUDED.required,
    plan_sha256 = EXCLUDED.plan_sha256
