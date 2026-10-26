SELECT platform.platform_async_release_job_retention_v1(
    $1::uuid,
    $2::uuid,
    $3::text,
    $4::bigint,
    $5::text,
    $6::bigint,
    $7::bigint,
    $8::text,
    $9::timestamptz
);
