SELECT analytics_publication_repair_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::bigint,$3::timestamptz
);
