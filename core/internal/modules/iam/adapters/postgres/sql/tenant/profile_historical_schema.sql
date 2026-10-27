SELECT iam_profile_historical_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid
);
