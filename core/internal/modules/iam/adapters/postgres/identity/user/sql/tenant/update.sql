SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_identity_user_update_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::bigint, $5::uuid, $6::bytea,
    $7::bytea, $8::uuid, $9::uuid, $10::bigint, $11::bigint
);
