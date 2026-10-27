SELECT receipt_id,package_id,package_slug,package_version,release_id,
       component_id,owner_generation,contract,content_sha256,
       component_kind,module_slug,semantic_identity,semantic_sha256,
       item_count,auxiliary_count,document_bytes,idempotency_key,
       request_sha256,lifecycle_state,ready_at,created_at,updated_at
FROM iam_catalog_project_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::uuid,$3::text,$4::text,$5::uuid,$6::uuid,$7::bigint,
    $8::text,$9::text,$10::text,$11::text,$12::text,$13::bytea,
    $14::integer,$15::integer,$16::bytea,$17::text,
    $18::bytea,$19::timestamptz
)
