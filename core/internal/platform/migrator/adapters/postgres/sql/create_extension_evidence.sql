INSERT INTO platform.extension_registry (
    name,
    version,
    schema_name,
    owner_role,
    control_file_sha256,
    capabilities,
    database_origin,
    verified_run_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6::text[],
    $7,
    $8
)
ON CONFLICT (name, version, verified_run_id) DO NOTHING;
