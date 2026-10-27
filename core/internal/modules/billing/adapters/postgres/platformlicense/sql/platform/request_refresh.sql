SELECT operation_id, replayed, created_at
FROM platform.billing_platform_license_refresh_request_v1(
    $1, $2, $3, $4, $5
);
