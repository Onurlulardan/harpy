SELECT authorization_clock, result_revision, replayed
FROM iam_permission_cache_flush_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::uuid,$3::bytea,$4::uuid,$5::bigint,$6::bigint,$7::bigint
)
