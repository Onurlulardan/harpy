SELECT meter_count, replay
FROM platform.billing_installation_open_usage_v2(
    $1::uuid, $2::uuid, $3::text
);
