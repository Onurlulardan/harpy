SELECT * FROM iam_identity_user_read_v3(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::uuid, $4::bigint, $5::bigint
);
