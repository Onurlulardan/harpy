SELECT iam_profile_document_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::text,
    $2::uuid
);
