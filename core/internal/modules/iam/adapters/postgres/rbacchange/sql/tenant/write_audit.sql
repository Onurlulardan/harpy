SELECT iam_rbac_audit_write_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::uuid, $4::bytea,
    $5::text, $6::text, $7::bigint, $8::bigint
);
