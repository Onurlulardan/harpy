SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_identity_user_relations_replace_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::text, $3::bytea, $4::uuid, $5::bigint,
    $6::boolean, $7::uuid, $8::bytea, $9::bytea,
    $10::boolean, $11::uuid[], $12::uuid[],
    $13::boolean, $14::uuid[], $15::uuid[],
    $16::uuid, $17::uuid, $18::bigint, $19::bigint
);
