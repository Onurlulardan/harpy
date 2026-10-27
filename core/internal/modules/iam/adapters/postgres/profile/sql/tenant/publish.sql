SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_profile_schema_publish_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::text, $5::bigint,
    $6::jsonb, $7::bytea, $8::uuid, $9::uuid, $10::uuid,
    $11::bigint, $12::bigint
);
