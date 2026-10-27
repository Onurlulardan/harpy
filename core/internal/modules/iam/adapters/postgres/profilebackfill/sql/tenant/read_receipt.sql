SELECT command_key, request_sha256, result_id,
       COALESCE(target_schema_id::text, ''),
       result_revision, authorization_clock
FROM iam_profile_publication_receipt_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::uuid, $4::bigint, $5::bigint
);
