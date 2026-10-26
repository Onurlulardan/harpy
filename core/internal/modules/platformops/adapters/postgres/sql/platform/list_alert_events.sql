SELECT id, environment_id, alert_rule_id, service_name, previous_status,
       current_status, channels_used, message, resolved_at, created_at, total_count
FROM platform.platformops_list_alert_events_v1(
    $1::uuid,$2::text,$3::text,$4::integer,$5::integer
)
