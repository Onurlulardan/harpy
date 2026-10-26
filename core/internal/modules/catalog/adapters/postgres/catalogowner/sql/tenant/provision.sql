SELECT receipt_id,package_id,package_slug,package_version,release_id,
       component_id,owner_generation,contract,content_sha256,
       module_slug,source_contract,semantic_sha256,
       form_count,row_count,document_bytes,idempotency_key,
       request_sha256,lifecycle_state,ready,retry_at,completed_at,
       created_at,updated_at
FROM catalog_demo_project_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::uuid,$3::text,$4::text,$5::uuid,$6::uuid,$7::bigint,
    $8::text,$9::text,$10::text,$11::text,$12::bytea,
    $13::integer,$14::integer,$15::bytea,$16::text,
    $17::bytea,$18::timestamptz
)
