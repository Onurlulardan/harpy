SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_service_principal_create_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::text, $5::text,
    $6::uuid, $7::uuid, $8::bigint, $9::bigint, $10::bytea,
    $11::timestamptz, $12::boolean, $13::boolean, $14::boolean
);
