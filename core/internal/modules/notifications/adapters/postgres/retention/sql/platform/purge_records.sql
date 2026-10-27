SELECT platform.notifications_retention_purge_platform_v1(
    $1::uuid, $2::timestamptz, $3::integer
);
