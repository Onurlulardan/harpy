SELECT platform.platform_async_finalize_stale_external_unknown_v1(
    $1::uuid,
    $2::uuid,
    $3::bigint,
    $4::uuid,
    $5::timestamptz
);
