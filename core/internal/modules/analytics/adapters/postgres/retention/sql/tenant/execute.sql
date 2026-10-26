SELECT * FROM analytics_execute_retention_cycle_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::uuid,$3::uuid,$4::uuid,$5,$6,$7::integer
);
