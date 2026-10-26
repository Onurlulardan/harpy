SELECT iam_profile_active_catalog_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::text
);
