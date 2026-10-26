SELECT resource_id::text,version_id::text,head_id::text,
       lifecycle_state,resource_reference_epoch
FROM analytics_catalog_owner_binding_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::uuid,$3::bigint,$4::text,$5::text
)
