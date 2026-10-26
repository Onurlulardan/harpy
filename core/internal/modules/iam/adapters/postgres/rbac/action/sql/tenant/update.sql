SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_rbac_action_update_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::text, $5::text, $6::bigint,
    $7::uuid, $8::uuid, $9::bigint, $10::bigint
);
