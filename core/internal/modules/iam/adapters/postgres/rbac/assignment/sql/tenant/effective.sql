SELECT * FROM iam_rbac_effective_permissions_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::text, $2::uuid, $3::integer, $4::uuid, $5::uuid
);
