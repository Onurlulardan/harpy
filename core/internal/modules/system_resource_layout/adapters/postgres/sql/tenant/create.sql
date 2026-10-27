SELECT system_resource_layout_create_v3(
    (pg_catalog.current_schemas(false))[2]::name, $1, $2, $3, $4, $5
);
