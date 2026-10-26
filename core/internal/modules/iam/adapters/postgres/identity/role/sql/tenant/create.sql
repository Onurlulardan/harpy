SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_identity_role_create_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::bytea, $4::uuid, $5::text, $6::uuid, $7::uuid,
    $8::text, $9::text, $10::boolean, $11::uuid, $12::bytea, $13::bytea,
    $14::uuid, $15::uuid, $16::bigint, $17::bigint
);
