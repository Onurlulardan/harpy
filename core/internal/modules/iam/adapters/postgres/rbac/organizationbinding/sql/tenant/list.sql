SELECT * FROM iam_rbac_organization_binding_list_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::text, $3::integer, $4::uuid, $5::uuid
);
