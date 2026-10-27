SELECT analytics_query_root_by_version_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1::uuid
);
