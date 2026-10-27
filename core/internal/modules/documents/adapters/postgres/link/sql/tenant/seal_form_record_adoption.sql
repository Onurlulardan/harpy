SELECT source_receipt_sha256, receipt_sha256, completed_at
FROM documents_form_record_adoption_seal_v1(
    $1::uuid,$2::uuid,$3::bigint,$4::uuid,$5::bigint,$6::bigint,
    $7::bytea,$8::bytea,$9::bytea,$10::timestamptz
);
