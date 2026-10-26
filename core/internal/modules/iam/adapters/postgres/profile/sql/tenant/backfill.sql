SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_profile_document_backfill_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::text, $4::uuid, $5::uuid, $6::uuid,
    $7::bytea, $8::bytea, $9::bigint, $10::uuid, $11::uuid,
    $12::bigint, $13::bigint
);
