SELECT platform.platform_async_park_stale_work_v2(
    $1::uuid,
    $2::text,
    $3::uuid,
    $4::bigint,
    $5::uuid
);
