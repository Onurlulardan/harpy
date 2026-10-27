SELECT analytics_public_dashboard_root_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,
    $2::uuid
);
