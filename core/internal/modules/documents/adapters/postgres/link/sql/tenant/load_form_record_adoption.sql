SELECT operation_id, definition_id, reference_epoch, cursor_record_id,
       record_count, link_count, source_catalog_sha256, link_catalog_sha256,
       lifecycle_state IN ('READY','SUCCEEDED')
FROM documents_form_record_adoption_load_v1(
    $1::uuid,$2::uuid,$3::bigint,$4::bytea,$5::bytea,$6::timestamptz
);
