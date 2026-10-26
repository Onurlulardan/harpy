SELECT source_id,resource_kind,format,plaintext_bytes,plaintext_sha256,
       uploaded_at,replayed
FROM iam_system_import_source_register_v2(
    (pg_catalog.current_schemas(false))[2]::name,$1::uuid,$2::uuid,$3::uuid,$4::text,
    $5::bigint,$6::bigint,$7::bytea,$8::text,$9::text,$10::text,$11::bytea,
    $12::bytea,$13::bigint,$14::bytea,$15::bigint,$16::text,$17::bytea,
    $18::bytea,$19::text,$20::timestamptz
)
