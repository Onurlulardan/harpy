SELECT query_id, version_id, version_no, root_revision, reference_epoch,
       version_revision, lifecycle_state, version_state, name, query_document,
       parameter_document, output_document, manifest_sha256, cache_policy,
       created_at, updated_at, replayed
FROM analytics_query_publish_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::text, $4::bytea, $5::uuid, $6::uuid, $7::uuid,
    $8::bigint, $9::bigint, $10::bytea, $11::bytea, $12::bytea,
    $13::jsonb, $14::bytea, $15::text, $16::timestamptz
);
