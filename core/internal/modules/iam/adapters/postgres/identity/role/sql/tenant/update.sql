SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_identity_role_update_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::bigint, $5::text, $6::text,
    $7::boolean, $8::uuid, $9::bytea, $10::bytea,
    $11::uuid, $12::uuid, $13::bigint, $14::bigint
);
