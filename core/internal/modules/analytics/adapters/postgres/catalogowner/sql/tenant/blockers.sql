SELECT receipt_id::text, operation_id::text, component_id::text,
       reference_epoch, catalog_version, zero_digest,
       blockers::text::bytea, completed_at
FROM analytics_catalog_component_blockers_v1(
    $1::uuid,$2::uuid,$3::uuid,$4::bigint,$5::bigint,$6::timestamptz
)
