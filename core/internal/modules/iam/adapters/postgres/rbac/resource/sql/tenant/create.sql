SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_rbac_resource_create_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::text, $5::text, $6::text,
    $7::text, $8::text, $9::boolean, $10::uuid, $11::uuid, $12::bigint, $13::bigint
);
