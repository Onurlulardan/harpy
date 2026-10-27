SELECT system_resource_layout_update_v2(
    (pg_catalog.current_schemas(false))[2]::name, $1, $2, $3, $4, $5
);
