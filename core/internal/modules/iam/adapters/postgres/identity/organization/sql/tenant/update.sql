SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_identity_organization_update_v3(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::bigint, $5::text, $6::text, $7::uuid,
    $8::boolean, $9::uuid[], $10::uuid[], $11::uuid, $12::bytea, $13::bytea,
    $14::uuid, $15::uuid, $16::bigint, $17::bigint
);
