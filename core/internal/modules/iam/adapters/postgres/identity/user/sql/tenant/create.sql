SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_identity_user_create_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::bytea, $4::uuid, $5::uuid, $6::uuid, $7::uuid,
    $8::uuid, $9::bytea, $10::bytea, $11::text, $12::text, $13::bytea,
    $14::text, $15::bytea, $16::uuid, $17::uuid, $18::bigint, $19::bigint
);
