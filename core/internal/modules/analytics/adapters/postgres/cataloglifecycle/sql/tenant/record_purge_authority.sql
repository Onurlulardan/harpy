SELECT operation_id::text,component_id::text,reference_epoch,
       authority_sha256,resource_id::text,version_id::text,
       head_id::text,recorded_at
FROM analytics_catalog_owner_purge_authority_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::uuid,$3::uuid,$4::bigint,$5::text,$6::text,
    $7::bigint,$8::text,$9::uuid,$10::uuid,$11::uuid,
    $12::timestamptz
)
