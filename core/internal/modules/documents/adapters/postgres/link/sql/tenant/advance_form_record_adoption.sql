SELECT operation_id, definition_id, reference_epoch, cursor_record_id,
       record_count, link_count, source_catalog_sha256, link_catalog_sha256,
       lifecycle_state IN ('READY','SUCCEEDED')
FROM documents_form_record_adoption_advance_v1(
    $1::uuid,$2::uuid,$3::bigint,$4::uuid,$5::bigint,$6::bigint,
    $7::bytea,$8::bytea,$9::uuid,$10::bigint,$11::bigint,
    $12::bytea,$13::bytea,$14::boolean,$15::timestamptz
);
