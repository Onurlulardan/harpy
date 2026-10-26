SELECT analytics_resource_manifest_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1::text, $2::uuid
);
