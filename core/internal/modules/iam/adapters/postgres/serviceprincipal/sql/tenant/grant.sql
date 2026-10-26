SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_capability_grant_write_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::uuid, $5::bigint,
    $6::text, $7::text, $8::text, $9::text, $10::uuid,
    $11::timestamptz, $12::timestamptz, $13::bytea,
    $14::uuid, $15::uuid, $16::bigint, $17::bigint, $18::bytea,
    $19::timestamptz, $20::boolean, $21::boolean, $22::boolean
);
