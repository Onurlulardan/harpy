SELECT blockers::text::bytea,resource_id::text,
       version_id::text,head_id::text
FROM analytics_catalog_owner_preflight_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::uuid,$3::bigint,$4::text,$5::text,$6::bigint
)
