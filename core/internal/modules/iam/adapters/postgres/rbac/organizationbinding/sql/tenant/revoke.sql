SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_rbac_organization_binding_revoke_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::uuid, $5::uuid, $6::bigint,
    $7::uuid, $8::uuid, $9::bigint, $10::bigint
);
