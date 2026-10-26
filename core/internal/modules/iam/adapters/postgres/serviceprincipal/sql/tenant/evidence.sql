SELECT iam_capability_grant_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,
    $2::bigint
);
