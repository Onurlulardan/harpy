SELECT handoff_id,provision_receipt_id,component_id,package_id,release_id,
       forms_run_id,
       request_sha256,form_count,row_count,readback_sha256,completed_at
FROM catalog_demo_record_handoff_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $2::bigint,$1::uuid,$3::uuid,$4::uuid,$5::uuid,$6::uuid,
    $7::uuid,$8::bytea,$9::integer,$10::integer,$11::bytea,$12::timestamptz
)
