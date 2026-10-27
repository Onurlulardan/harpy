SELECT *
FROM iam_authorization_evidence_read_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,
    $2::uuid,
    $3::text,
    $4::text
);

