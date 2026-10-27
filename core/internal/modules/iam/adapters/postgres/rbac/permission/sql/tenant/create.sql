SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_rbac_permission_create_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::text, $5::uuid, $6::uuid,
    $7::boolean, $8::uuid, $9::uuid, $10::bigint, $11::bigint
);
