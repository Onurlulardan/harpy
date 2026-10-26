SELECT * FROM iam_identity_role_list_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::integer, $3::uuid, $4::uuid, $5::bigint, $6::bigint
);
