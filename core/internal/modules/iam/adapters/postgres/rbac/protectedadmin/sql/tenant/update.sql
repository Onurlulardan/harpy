SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_rbac_protected_admin_update_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::integer, $5::bigint,
    $6::uuid, $7::uuid, $8::bigint, $9::bigint
);
