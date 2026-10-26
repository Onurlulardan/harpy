SELECT id, environment_id, alert_rule_id, service_name, previous_status,
       current_status, channels_used, message, resolved_at, created_at
FROM platform.platformops_test_alert_rule_v1(
    $1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::text,$6::timestamptz
)
