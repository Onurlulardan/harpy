SELECT system_resource_layout_list_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1
);
