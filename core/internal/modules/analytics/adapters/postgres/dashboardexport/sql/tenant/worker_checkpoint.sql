SELECT lease_until, cancellation_requested
FROM analytics_dashboard_export_platform_checkpoint_v1(
    $1::name,$2,$3,$4,$5,$6,$7,$8
)
