SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_identity_organization_create_v3(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::text, $5::text, $6::uuid,
    $7::uuid[], $8::uuid[], $9::uuid, $10::bytea, $11::bytea,
    $12::uuid, $13::uuid, $14::bigint, $15::bigint
);
