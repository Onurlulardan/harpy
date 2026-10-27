SELECT analytics_dashboard_set_default_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::text, $4::bytea, $5::uuid,
    $6::uuid, $7::text, $8::timestamptz
);
